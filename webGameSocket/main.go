package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var temp map[string]*template.Template
var users []User
var clients map[*client]bool
var startgame bool
var startCount = 4
var words = []string{"dropât", "mirons", "lumens", "amples", "meulez", "filées", "userai", "mégota", "battit", "entuba", "cuvées", "rances", "liftes", "boueux", "frotte", "parvis", "écriez", "obséda", "dianes", "fédéré", "piffés", "vendre", "tincal", "gringe", "perdît", "fâcher", "baissé", "langui", "dorent", "dévoue", "adjugé", "paillé", "payais", "rapina", "truite", "adapté", "grands", "frirez", "luises", "chérît", "tordît", "rejoué", "sassés", "cosmos", "arceau", "mollah", "louvas", "irions", "fripon", "amande", "blesse", "serins", "mariés", "tendît", "rocker", "prison", "chauve", "montra", "moirât", "grigna", "airent", "gommât", "zingue", "caille", "classa", "guéait", "values", "ombrât", "repart", "loupai", "balaya", "lavais", "brûlez", "draves", "barrât", "gallon", "criait", "lapent", "emmura", "flouée", "frappe", "poquas", "alitas", "épices", "viseux", "scions", "clapit", "xystes", "boulée", "frêles", "linéal", "crèche", "alitât", "pincer", "dodues", "dégota", "birème", "excite", "tachas", "puffin", "bourda", "peseta", "ormoie", "parque", "tissés", "gésier", "renais", "geints", "marrai", "paraît", "dévide", "égermé", "gavial", "gattes", "roguée", "mazout", "alloué", "ruchez", "étoupé", "épeure", "tillât", "épuise", "nouvel", "anisée", "noueux", "santon", "apures", "balisa", "épeler", "pannés", "greffé", "ramera", "dorait", "épeire", "jaspai", "étêtai", "veiner", "élimai", "médisé", "tançât", "pulsas", "bannez", "déniai", "décape", "vengea", "moisas", "écarté", "publie", "tarait", "zazous"}
var rw string     // selected word
var userIndex = 0 // the player start index
var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type dataTemplate struct {
	Users        []User   `json:"users"`
	UserNotReady []User   `json:"user_not_ready"`
	StartGame    bool     `json:"start_game"`
	TimeTick     int      `json:"time_tick"`
	RandWord     []string `json:"choose_word"`
}

type client struct {
	user        *User
	conn        *websocket.Conn
	mu          sync.Mutex
	broadcast   chan serverEvent
	register    chan *client
	unregister  chan serverEvent
	ready       chan serverEvent
	clientEvent *clientEvent
}

type User struct {
	Id      string `json:"id"`
	Pseudo  string `json:"pseudo"`
	Status  string `json:"status"`
	Logo    []byte
	Turn    string `json:"turn"`
	Current bool   `json:"current"`
	Win     bool
	Score   int `json:"score"`
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

	mux := http.NewServeMux()
	//t, err := newTemplateCache("./ui/html/")
	t, err := newTemplateCache("/Users/phenix/home/go/src/GuessGame/webGameSocket/ui/html")
	if err != nil {
		fmt.Println(err)
	}

	temp = t

	clients = make(map[*client]bool)

	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		render(writer, request, "login.page.tmpl", nil)
	})

	mux.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {

		conn, err := upgrade.Upgrade(writer, request, nil)
		if err != nil {
			fmt.Println(err)
			conn.Close()
			return
		}

		createNewSocketClient(conn)

	})

	//fileServer := http.FileServer(http.Dir("./ui/static/"))
	fileServer := http.FileServer(http.Dir("/Users/phenix/home/go/src/GuessGame/webGameSocket/ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	fmt.Println("starting server at : 4000 port")
	log.Fatalln(http.ListenAndServe(":4000", mux))
}

func unAutorizeClient(c *client) {
	c.user.Status = "rejected"
	c.broadcast <- serverEvent{
		SocketID: c.user.Id,
		Template: renderToString("login.page.tmpl", c.user),
		Status:   "rejected",
	}

	fmt.Println(c)
}

func createNewSocketClient(conn *websocket.Conn) {
	cli := &client{
		user: &User{
			Id:     uuid.New().String(),
			Status: "join",
			Turn:   "wait",
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
		ready:      make(chan serverEvent),
	}

	fmt.Println("client joined session id => ", cli.user.Id)
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

func getusers() (u []User) {

	for c := range clients {
		u = append(u, *c.user)
	}
	return u
}

func initTurnToWait() {
	for c := range clients {
		c.user.Turn = "wait"
	}
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
		case se, ok := <-client.ready:
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
			//fmt.Println("im in case tikcet.C")
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}

}

func handleLoginClient(c *client) {
	c.user.Pseudo = c.clientEvent.Payload
	users = getusers()
	ResponseToAllUsers("gamearea.page.tmpl", "broadcast", dataTemplate{
		Users:        users,
		UserNotReady: nil,
		StartGame:    startgame,
		TimeTick:     startCount,
	})
}

func readyUser(c *client) {
	c.user.Status = "ready"
	// check if we can start th game
	startgame = true

	// iterate userIndex if one of then not ready don't start game
	for c := range clients {
		if c.user.Status != "ready" {
			startgame = false
		}
	}

	if startgame {
		ran := rand.Intn(len(users))
		users[ran].Turn = "start"

	}
	ResponseToAllUsers("gamearea.page.tmpl", "ready", dataTemplate{
		Users:        users,
		UserNotReady: nil,
		StartGame:    startgame,
		TimeTick:     startCount,
	})

	if startgame {
		go timeTick()
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
			if len(clients) == 5 {
				unAutorizeClient(client)
				return
			}
			handleLoginClient(client)
		case "closed":
			fmt.Println("im here too closed")
			unRegisterAndCloseConnection(client)
		case "error":
			fmt.Println("im here too in error")
			unRegisterAndCloseConnection(client)
		case "ready":
			readyUser(client)
		case "guess":
			fmt.Println("coucou")
			checkGuess(client)
		}

	}

}

func checkGuess(client *client) {

	guess := client.clientEvent.Payload

	if checkWord(guess, rw) {
		client.user.Score++
		userIndex++
		users := getusers()
		ResponseToAllUsers("gamearea.page.tmpl", "guess", dataTemplate{
			Users:        users,
			UserNotReady: nil,
			StartGame:    startgame,
			TimeTick:     startCount,
		})
	} else {
		cw := choosedWord(words)
		userIndex++
		users := getusers()
		ResponseToAllUsers("gamearea.page.tmpl", "guess", dataTemplate{
			Users:        users,
			UserNotReady: nil,
			StartGame:    startgame,
			TimeTick:     startCount,
			RandWord:     cw,
		})
	}
}

func unRegisterAndCloseConnection(client *client) {
	client.user.Status = "disconnect"
	users := getusers()
	ResponseToAllUsers("gamearea.page.tmpl", "unregister", dataTemplate{
		Users:        users,
		UserNotReady: nil,
		StartGame:    startgame,
		TimeTick:     startCount,
	})
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

func timeTick() {
	c := time.Tick(time.Second * 1)
myForTick:
	for {
		select {
		case <-c:
			startCount--
			ResponseToAllUsers("gamearea.page.tmpl", "ready", dataTemplate{
				Users:        users,
				UserNotReady: nil,
				StartGame:    startgame,
				TimeTick:     startCount,
			})
			if startCount == 0 {
				//send word to users
				cw := choosedWord(words)
				ResponseToAllUsers("gamearea.page.tmpl", "guess", dataTemplate{
					Users:        users,
					UserNotReady: nil,
					StartGame:    startgame,
					TimeTick:     startCount,
					RandWord:     cw,
				})
				break myForTick
			}
		}
	}

}

func ResponseToAllUsers(temp string, stat string, dt dataTemplate) {
	for l := range clients {
		se := serverEvent{
			SocketID: l.user.Id,
			Template: renderToString(temp, dt),
			Status:   l.user.Status,
		}
		switch stat {
		case "ready":
			l.ready <- se
		case "unregister":
			l.unregister <- se
		case "broadcast":
			l.broadcast <- se
		case "guess":
			l.broadcast <- se
		}
	}
}

func choosedWord(m []string) []string {
	ran := rand.Intn(len(words))
	rw = words[ran]                               // rw => random world pickup
	words = append(words[:ran], words[ran+1:]...) // delete the selected word
	sow := strings.Split(rw, "")                  // sow => slice of caracters from word
	return MixWord2(sow)                          // mixed word
}

func MixWord2(word []string) []string {

	var res []string
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for _, i := range r.Perm(len(word)) {
		val := word[i]
		res = append(res, val)
	}
	return res
}

func checkWord(guess string, w string) bool {
	if strings.ToLower(guess) == strings.ToLower(w) {
		return true
	}
	return false
}

func randUserToPlayFirst() {

}
