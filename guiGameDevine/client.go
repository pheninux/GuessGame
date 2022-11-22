//client.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Max wait time when writing message to peer
	writeWait = 10 * time.Second

	// Max time till next pong from peer
	pongWait = 60 * time.Second

	// Send ping interval, must be less then pong wait time
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 10000
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type socketEvent struct {
	EventName    string       `json:"event_name"`
	EventPayload interface{}  `json:"event_payload"`
	DataTemplate DataTemplate `json:"data_template"`
}

// Client represents the websocket client at the server
type Client struct {
	// The actual websocket connection.
	Conn     *websocket.Conn  `json:"-"`
	WsServer *WsServer        `json:"-"`
	Send     chan socketEvent `json:"-"`
	Id       string           `json:"id"`
}

func newClient(conn *websocket.Conn, wsServer *WsServer) *Client {

	return &Client{
		Conn:     conn,
		WsServer: wsServer,
		Send:     make(chan socketEvent),
		Id:       uuid.New().String(),
	}
}

// ServeWs handles websocket requests from clients requests.
func (app *application) ServeWs(wsServer *WsServer, w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := newClient(conn, wsServer)

	app.WsServer.clients = map[*Client]bool{client: true}

	go client.writePump()
	go client.readPump()

	temp := app.renderPartialToString("./ui/players.partial.tmpl", app.players)
	fmt.Println("template => ", temp)

	var se socketEvent = socketEvent{
		EventName:    "login",
		EventPayload: temp,
		DataTemplate: DataTemplate{Client: client},
	}
	fmt.Println(se)
	wsServer.register <- se
}

func (client *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.Conn.Close()

	}()
	//for {
	//	select {
	//	case message, ok := <-client.Send:
	//		client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
	//		if !ok {
	//			// The WsServer closed the channel.
	//			client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
	//			return
	//		}
	//
	//		w, err := client.Conn.NextWriter(websocket.TextMessage)
	//		if err != nil {
	//			return
	//		}
	//		w.Write(message)
	//
	//		// Attach queued chat messages to the current websocket message.
	//		n := len(client.Send)
	//		for i := 0; i < n; i++ {
	//			w.Write(newline)
	//			w.Write(<-client.Send)
	//		}
	//
	//		if err := w.Close(); err != nil {
	//			return
	//		}
	//	case <-ticker.C:
	//		client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
	//		if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
	//			return
	//		}
	//	}
	//}

	for {
		select {
		case payload, ok := <-client.Send:
			reqBodyBytes := new(bytes.Buffer)
			json.NewEncoder(reqBodyBytes).Encode(payload)
			finalPayload := reqBodyBytes.Bytes()

			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write(finalPayload)

			n := len(client.Send)
			for i := 0; i < n; i++ {
				json.NewEncoder(reqBodyBytes).Encode(<-client.Send)
				w.Write(reqBodyBytes.Bytes())
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}

}

func (client Client) readPump() {
	defer func() {
		client.disconnect()
	}()

	client.Conn.SetReadLimit(maxMessageSize)
	client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetPongHandler(func(string) error { client.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// Start endless read loop, waiting for messages from client
	for {
		_, jsonMessage, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v", err)
			}
			break
		}
		var se socketEvent
		err = json.Unmarshal(jsonMessage, &se)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("data recieved", se.EventPayload)

		switch se.EventName {
		case "login":
			fmt.Println("login switch")
			client.WsServer.broadcast <- se
		}

	}
}
func (client *Client) disconnect() {
	client.WsServer.unregister <- socketEvent{
		EventName:    "disconnect",
		EventPayload: nil,
		DataTemplate: DataTemplate{client},
	}
	close(client.Send)
	client.Conn.Close()
}
