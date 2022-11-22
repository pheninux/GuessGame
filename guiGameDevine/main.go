package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
)

type players struct {
	Pseudo string
	Id     string
	ws     *websocket.Conn
}
type application struct {
	temp     map[string]*template.Template
	players  []players
	ws       *websocket.Conn
	WsServer *WsServer
}

func main() {
	temp, err := newTemplateCache("./ui/")
	//temp, err := newTemplateCache("C:\\Users\\a706836\\go\\src\\DevineGame\\guiGameDevine\\ui\\")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(temp)
	wsServer := NewWebsocketServer()

	app := &application{temp: temp, WsServer: wsServer}

	go app.Run()
	//go func() {
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
	//
	//}()
	fmt.Println("server start at : 4000 port")
	log.Fatalln(http.ListenAndServe(":4000", app.routes()))
}
