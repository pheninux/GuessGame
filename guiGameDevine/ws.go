package main

//func (app *application) wsLoginForm(w http.ResponseWriter, r *http.Request) {
//
//	// upgrade this connection to a WebSocket
//	// connection
//	ws, err := upgrader.Upgrade(w, r, nil)
//
//	if err != nil {
//		log.Println(err)
//	}
//
//	err = ws.WriteMessage(1, app.renderByte(w, r, "login.page.tmpl", app.players))
//	if err != nil {
//		log.Println(err)
//	}
//}
//func (app *application) wsLogin(w http.ResponseWriter, r *http.Request) {
//
//	// upgrade this connection to a WebSocket
//	// connection
//	ws, _ := upgrader.Upgrade(w, r, nil)
//
//	// read in a message
//	_, p, err := ws.ReadMessage()
//	if err != nil {
//		log.Println(err)
//		ws.Close()
//		return
//	}
//	// print out that message for clarity
//	fmt.Println(string(p))
//	if string(p) != "" {
//		chanPlayer <- string(p)
//	}
//
//	for {
//		select {
//		case p := <-chanPlayer:
//			fmt.Println("player subsribe => ", p)
//			app.players = append(app.players, struct{ Pseudo string }{Pseudo: p})
//			err := app.ws.WriteMessage(1, app.renderByte2("gameArea.page.tmpl", app.players))
//			if err != nil {
//				log.Println(err)
//				app.ws.Close()
//				return
//			}
//		}
//	}
//}
//
//// define a reader which will listen for
//// new messages being sent to our WebSocket
//// endpoint
//func (app *application) reader(conn *websocket.Conn, w http.ResponseWriter, r *http.Request) {
//	for {
//		// read in a message
//		messageType, p, err := conn.ReadMessage()
//		if err != nil {
//			log.Println(err)
//			conn.Close()
//			return
//		}
//		// print out that message for clarity
//		fmt.Println(string(p))
//
//		app.players = append(app.players, struct{ Pseudo string }{Pseudo: string(p)})
//		fmt.Println(app.players)
//		data := app.renderByte(w, r, "gameArea.page.tmpl", app.players)
//		if err := conn.WriteMessage(messageType, data); err != nil {
//			log.Println(err)
//			conn.Close()
//			return
//		}
//	}
//}
