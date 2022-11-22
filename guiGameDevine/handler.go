package main

import (
	"github.com/google/uuid"
	"net/http"
)

func (app *application) home(writer http.ResponseWriter, request *http.Request) {
	http.Redirect(writer, request, "/login", 302)
}

func (app *application) login(writer http.ResponseWriter, request *http.Request) {
	app.render(writer, request, "login.page.tmpl", nil)
}
func (app *application) wsHandler(w http.ResponseWriter, r *http.Request) {
	app.players = append(app.players, players{Pseudo: r.PostForm.Get("pseudo"), Id: uuid.New().String()})
	app.ServeWs(app.WsServer, w, r)
}

func (app *application) gameArea(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	app.players = append(app.players, players{Pseudo: r.PostForm.Get("pseudo"), Id: uuid.New().String()})
	app.render(w, r, "gameArea.page.tmpl", app.players)
}

//todo il faut pouvoir identifier chaque cklien par sa connexion pour un envoie multiple
