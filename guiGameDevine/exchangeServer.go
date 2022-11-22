package main

type DataTemplate struct {
	Client *Client
}

type WsServer struct {
	clients    map[*Client]bool
	register   chan socketEvent
	unregister chan socketEvent
	broadcast  chan socketEvent
}

// NewWebsocketServer creates a new WsServer type
func NewWebsocketServer() *WsServer {
	return &WsServer{
		clients:    make(map[*Client]bool),
		register:   make(chan socketEvent),
		unregister: make(chan socketEvent),
		broadcast:  make(chan socketEvent),
	}
}

// Run our websocket server, accepting various requests
func (app *application) Run() {
	for {
		select {
		case se := <-app.WsServer.register:
			app.registerClient(se)
		case client := <-app.WsServer.unregister:
			app.unregisterClient(client)
		case message := <-app.WsServer.broadcast:
			app.broadcastToClients(message)
		}
	}

}

func (app *application) registerClient(se socketEvent) {

	app.notifyOtherClient(se)
}

func (app *application) notifyOtherClient(se socketEvent) {

	//b, err := json.Marshal(se)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//fmt.Println("client marshaled => ", b)
	//fmt.Println("data template => ", se)
	//fmt.Println(app.WsServer.clients)

	for client := range app.WsServer.clients {
		client.Send <- se
	}

}

func (app *application) unregisterClient(se socketEvent) {
	if _, ok := app.WsServer.clients[se.DataTemplate.Client]; ok {
		delete(app.WsServer.clients, se.DataTemplate.Client)
	}
}

func (app *application) broadcastToClients(se socketEvent) {
	for client := range app.WsServer.clients {
		client.Send <- se
	}
}
