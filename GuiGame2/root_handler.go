package main

import (
	"fmt"
	"github.com/google/uuid"
	"net/http"
)

func (app *application) routes() http.Handler {

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/login", app.loginTmpl)
	mux.HandleFunc("/gameArea", app.gameArea)
	mux.HandleFunc("/wsHandler", app.wslogin)

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	return mux

}

func (app *application) home(writer http.ResponseWriter, request *http.Request) {
	http.Redirect(writer, request, "/login", 302)
}

func (app *application) loginTmpl(writer http.ResponseWriter, request *http.Request) {
	app.render(writer, request, "login.page.tmpl", nil)
}
func (app *application) wslogin(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
	}
	// Reading username from request parameter
	username := r.URL.Query().Get(":pseudo")
	//create new client
	client := newClient(conn, uuid.New(), username)
	//add to wsManager
	app.ws.client[client] = true

	client.dataTemplate = app.wrapDataTemplate(client, "login", app.renderPartialToString("./ui/players.partial.tmpl", "players", wrapUsersFromMap(app.ws.client)))

	app.client = client

	go app.writePump()
	go app.readPump()

	app.ws.register <- client
}

func (app *application) gameArea(w http.ResponseWriter, r *http.Request) {

}
