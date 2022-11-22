package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var clients map[*client]bool

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type client struct {
	user        user
	conn        *websocket.Conn
	mu          sync.Mutex
	broadcast   chan serverEvent
	register    chan *client
	unregister  chan serverEvent
	clientEvent *clientEvent
}

type user struct {
	id     string `json:"id"`
	pseudo string `json:"pseudo"`
	status string `json:"status"`
}

type clientEvent struct {
	Event   string `json:"event"`
	Payload string `json:"payload"`
}

type serverEvent struct {
	SocketID string `json:"socket_id"`
	Template string `json:"template"`
	Status   string `json:"status"`
}

func main() {

	clients = make(map[*client]bool)
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, "C:\\Users\\a706836\\go\\src\\DevineGame\\webGameSocket\\login.tmpl")
	})

	http.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {

		conn, err := upgrade.Upgrade(writer, request, nil)
		if err != nil {
			fmt.Println(err)
			conn.Close()
			return
		}

		createNewSocketClient(conn)

	})

	fileServer := http.FileServer(http.Dir("C:\\Users\\a706836\\go\\src\\DevineGame\\webGameSocket\\static"))
	http.Handle("/static/", http.StripPrefix("/static", fileServer))

	fmt.Println("starting server at : 4000 port")
	log.Fatalln(http.ListenAndServe(":4000", nil))
}

func createNewSocketClient(conn *websocket.Conn) {
	cli := &client{
		user: user{
			id:     uuid.New().String(),
			status: "join",
		},
		clientEvent: &clientEvent{
			Event:   "join",
			Payload: "",
		},
		conn:       conn,
		mu:         sync.Mutex{},
		broadcast:  make(chan serverEvent),
		unregister: make(chan serverEvent),
		register:   make(chan *client),
	}

	fmt.Println("client joined session id => ", cli.user.id)
	clients[cli] = true
	fmt.Println("clients in session ", clients)
	go cli.writer()
	go cli.reader()

}

func renderTemplate(fileDir string, data interface{}) string {
	buff := new(bytes.Buffer)
	t, err := template.ParseFiles(fileDir)
	if err != nil {
		fmt.Println(err)
	}
	if err := t.Execute(buff, data); err != nil {
		fmt.Println(err)
	}
	return buff.String()
}

func getusers() (u []user) {
	for c := range clients {
		u = append(u, c.user)
		fmt.Println(c.user)
	}
	fmt.Println("users => ", u)
	return u
}

func (client *client) writer() {

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()
	for {

		select {
		case se, ok := <-client.broadcast:
			err := client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				fmt.Println(err)
			}
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := client.conn.WriteJSON(se); err != nil {
				fmt.Println(err)
			}
		case se, ok := <-client.unregister:
			err := client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				fmt.Println(err)
			}
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := client.conn.WriteJSON(se); err != nil {
				fmt.Println(err)
			}
		case <-ticker.C:
			fmt.Println("im in case tikcet.C")
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}

}

func handleLoginClient(c *client) {
	c.user.pseudo = c.clientEvent.Payload
	for l := range clients {
		se := serverEvent{
			SocketID: l.user.id,
			Template: renderTemplate("C:\\Users\\a706836\\go\\src\\DevineGame\\webGameSocket\\gamearea.tmpl", getusers()),
			Status:   "join",
		}
		l.broadcast <- se
	}

}

func (client *client) reader() {

	var se clientEvent
	defer unRegisterAndCloseConnection(client)
	setSocketPayloadReadConfig(client)
	for {
		_, b, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error ===: %v", err)
			}
			break
		}
		json.Unmarshal(b, &se)
		client.clientEvent = &se
		switch se.Event {
		case "login":
			handleLoginClient(client)
		case "closed":
			fmt.Println("im here too closed")
			unRegisterAndCloseConnection(client)
		case "error":
			fmt.Println("im here too in error")
			unRegisterAndCloseConnection(client)
		case "test2":
			fmt.Println("test event fired from => ", client.user.pseudo)
		}

	}

}

func unRegisterAndCloseConnection(client *client) {
	client.user.status = "disconnect"

	for l := range clients {
		se := serverEvent{
			SocketID: l.user.id,
			Template: renderTemplate("C:\\Users\\a706836\\go\\src\\DevineGame\\webGameSocket\\gamearea.tmpl", getusers()),
			Status:   l.user.status,
		}
		l.unregister <- se
	}
	delete(clients, client)
	client.conn.Close()
}

func handleJoinClient(c *client) {
	fmt.Println("client joined session")
}

func setSocketPayloadReadConfig(c *client) {
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
}
