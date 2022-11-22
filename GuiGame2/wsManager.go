package main

func newWsManager() *wsManager {
	return &wsManager{
		client:        make(map[*client]bool),
		register:      make(chan *client),
		unregistred:   make(chan *client),
		broadcastchan: make(chan SocketEventStruct),
	}
}

func (app *application) run() {

	for {
		select {
		case client := <-app.ws.register:
			app.notifyUsers(client)

			//case client := <-app.ws.broadcastchan:
			//	app.notifyUsers(client)
		}

	}
}

func (app *application) notifyUsers(client *client) {
	for client := range app.ws.client {
		client.send <- client.dataTemplate
	}
}
