package main

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
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

func newClient(conn *websocket.Conn, uuid uuid.UUID, pseudo string) *client {
	return &client{
		conn: conn,
		user: &user{
			Pseudo: pseudo,
			ID:     uuid.String(),
		},
		send: make(chan *dataTemplate),
	}
}
func (app *application) writePump() {
	app.client.mu.Lock()
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		app.client.conn.Close()
		app.client.mu.Unlock()
	}()
	for {
		select {
		case payload, ok := <-app.client.send:
			reqBodyBytes := new(bytes.Buffer)
			json.NewEncoder(reqBodyBytes).Encode(payload)
			finalPayload := reqBodyBytes.Bytes()

			app.client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				app.client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := app.client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(finalPayload)

			n := len(app.client.send)
			for i := 0; i < n; i++ {
				json.NewEncoder(reqBodyBytes).Encode(<-app.client.send)
				w.Write(reqBodyBytes.Bytes())
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			app.client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := app.client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (app *application) readPump() {
	var socketEventPayload SocketEventStruct
	defer func() {
		app.client.disconnect()
	}()

	app.client.conn.SetReadLimit(maxMessageSize)
	app.client.conn.SetReadDeadline(time.Now().Add(pongWait))
	app.client.conn.SetPongHandler(func(string) error { app.client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, payload, err := app.client.conn.ReadMessage()

		decoder := json.NewDecoder(bytes.NewReader(payload))
		decoderErr := decoder.Decode(&socketEventPayload)

		if decoderErr != nil {
			log.Printf("error: %v", decoderErr)
			break
		}

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error ===: %v", err)
			}
			break
		}

		app.ws.broadcastchan <- socketEventPayload
	}
}

func (client *client) disconnect() {

}

//func handleSocketPayloadEvents(client *client, socketEventPayload SocketEventStruct) {
//	var socketEventResponse SocketEventStruct
//	switch socketEventPayload.EventName {
//	case "join":
//		log.Printf("Join Event triggered")
//		BroadcastSocketEventToAllClient(client.hub, SocketEventStruct{
//			EventName: socketEventPayload.EventName,
//			EventPayload: JoinDisconnectPayload{
//				UserID: client.userID,
//				Users:  getAllConnectedUsers(client.hub),
//			},
//		})
//
//	case "disconnect":
//		log.Printf("Disconnect Event triggered")
//		BroadcastSocketEventToAllClient(client.hub, SocketEventStruct{
//			EventName: socketEventPayload.EventName,
//			EventPayload: JoinDisconnectPayload{
//				UserID: client.userID,
//				Users:  getAllConnectedUsers(client.hub),
//			},
//		})
//
//	case "message":
//		log.Printf("Message Event triggered")
//		selectedUserID := socketEventPayload.EventPayload.(map[string]interface{})["userID"].(string)
//		socketEventResponse.EventName = "message response"
//		socketEventResponse.EventPayload = map[string]interface{}{
//			"username": getUsernameByUserID(client.hub, selectedUserID),
//			"message":  socketEventPayload.EventPayload.(map[string]interface{})["message"],
//			"userID":   selectedUserID,
//		}
//		EmitToSpecificClient(client.hub, socketEventResponse, selectedUserID)
//	}
//}
//
//func getUsernameByUserID(hub *Hub, userID string) string {
//	var username string
//	for client := range hub.clients {
//		if client.userID == userID {
//			username = client.username
//		}
//	}
//	return username
//}
//
//func getAllConnectedUsers(hub *Hub) []UserStruct {
//	var users []UserStruct
//	for singleClient := range hub.clients {
//		users = append(users, UserStruct{
//			Username: singleClient.username,
//			UserID:   singleClient.userID,
//		})
//	}
//	return users
//}
//
//// BroadcastSocketEventToAllClient will emit the socket events to all socket users
//func BroadcastSocketEventToAllClient(hub *wsManager, payload SocketEventStruct) {
//	for client := range hub.clients {
//		select {
//		case client.send <- payload:
//		default:
//			close(client.send)
//			delete(hub.clients, client)
//		}
//	}
//
