package main

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type client struct {
	Con      *websocket.Conn
	ID       string
	Register chan interface{}
	Sender   chan interface{}
}

type ServerEvent struct {
	Even string      `json:"even"`
	ID   string      `json:"id"`
	Data interface{} `json:"data"`
}

var clients = make(map[*client]bool)
var UUIDCli = make(map[string]*client)

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {

		t, err := template.ParseFiles("/Users/phenix/home/go/src/GuessGame/webSocketTraining/index.html")
		if err != nil {
			fmt.Println(err)
		}
		buff := new(bytes.Buffer)
		t.Execute(buff, nil)
		buff.WriteTo(writer)
	})

	mux.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {

		conn, err := upgrade.Upgrade(writer, request, nil)
		if err != nil {
			fmt.Println(err)
			conn.Close()
			return
		}

		id := request.URL.Query().Get("id")
		c, ok := UUIDCli[id]
		fmt.Println(c)

		if !ok {
			c = &client{
				Con:      conn,
				ID:       uuid.New().String(),
				Register: make(chan interface{}),
			}
			clients[c] = true
			UUIDCli[c.ID] = c

			go c.writer()
			go c.reader()
		}

		c.Register <- struct{}{}
	})

	//fileServer := http.FileServer(http.Dir("./ui/static/"))
	//fileServer := http.FileServer(http.Dir("/Users/phenix/home/go/src/GuessGame/webGameSocket/ui/static/"))
	//mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	fmt.Println("starting server at : 4000 port")
	log.Fatalln(http.ListenAndServe(":4000", mux))
}

func (client *client) writer() {

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		fmt.Printf("client ID : %s disconnected \n", client.ID)
		ticker.Stop()
		client.Con.Close()
	}()
	for {

		select {
		case _, ok := <-client.Register:
			err := client.Con.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				fmt.Println(err)
			}
			if !ok {
				client.Con.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			for v := range clients {
				se := ServerEvent{
					Even: "login",
					ID:   client.ID,
					Data: getConnectedClient(),
				}
				if err := v.Con.WriteJSON(se); err != nil {
					fmt.Println(err)
				}
			}

		case <-ticker.C:
			//fmt.Println("im in case tikcet.C")
			client.Con.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Con.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}

}

func (client *client) reader() {

	defer func() {
		fmt.Printf("client ID : %s disconnected \n", client.ID)
		client.Con.Close()
	}()
	for {
		_, b, err := client.Con.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error ===: %v", err)
			}
			break
		}
		switch string(b) {
		case "closed":
			fmt.Println("im here too closed")

		}
	}

}

func getConnectedClient() interface{} {
	var cliUuid []string
	for v := range clients {
		cliUuid = append(cliUuid, v.ID)
	}
	return cliUuid
}
