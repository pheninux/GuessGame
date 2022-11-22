package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type application struct {
	client       *client
	ws           *wsManager
	tmplFiles    map[string]*template.Template
	dataTemplate dataTemplate
}

func main() {

	// start scanning tmpl files

	temp, err := newTemplateCache("./ui/")
	//temp, err := newTemplateCache("C:\\Users\\a706836\\go\\src\\DevineGame\\guiGameDevine\\ui\\")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(temp)

	app := &application{
		ws:        newWsManager(),
		tmplFiles: temp,
	}

	//run ws manager
	go app.run()

	fmt.Println("starting server")
	log.Fatalln(http.ListenAndServe(":4000", app.routes()))
}
