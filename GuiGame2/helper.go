package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

//func (app *application) renderByte(w http.ResponseWriter, r *http.Request, name string, td interface{}) []byte {
//	// Retrieve the appropriate template set from the cache based on the page name
//	// (like 'home.page.tmpl'). If no entry exists in the cache with the
//	// provided name, call the serverError helper method that we made earlier.
//	ts, ok := app.temp[name]
//	if !ok {
//		http.Error(w, http.StatusText(404), 404)
//	}
//	// Initialize a new buffer.
//	buff := new(bytes.Buffer)
//	err := ts.Execute(buff, td)
//	if err != nil {
//		http.Error(w, http.StatusText(404), 404)
//	}
//
//	return buff.Bytes()
//}

//func (app *application) renderByte2(name string, td interface{}) []byte {
//	// Retrieve the appropriate template set from the cache based on the page name
//	// (like 'home.page.tmpl'). If no entry exists in the cache with the
//	// provided name, call the serverError helper method that we made earlier.
//	ts, ok := app.temp[name]
//	if !ok {
//		fmt.Println(errors.New("template introuvable !!"))
//	}
//	// Initialize a new buffer.
//	buff := new(bytes.Buffer)
//	err := ts.Execute(buff, td)
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	return buff.Bytes()
//}

//func (app *application) renderPartialToArrByte(name string, td interface{}) []byte {
//	// Retrieve the appropriate template set from the cache based on the page name
//	// (like 'home.page.tmpl'). If no entry exists in the cache with the
//	// provided name, call the serverError helper method that we made earlier.
//	ts, ok := app.temp[name]
//	if !ok {
//		fmt.Println(errors.New("template introuvable !!"))
//	}
//	// Initialize a new buffer.
//	buff := new(bytes.Buffer)
//	err := ts.ExecuteTemplate(buff, "players.partial.tmpl", td)
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	return buff.Bytes()
//}

func (app *application) render(w http.ResponseWriter, r *http.Request, name string, td interface{}) {
	// Retrieve the appropriate template set from the cache based on the page name
	// (like 'home.page.tmpl'). If no entry exists in the cache with the
	// provided name, call the serverError helper method that we made earlier.
	ts, ok := app.tmplFiles[name]
	if !ok {
		http.Error(w, http.StatusText(404), 404)
	}
	// Initialize a new buffer.
	buff := new(bytes.Buffer)
	err := ts.Execute(buff, td)
	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	buff.WriteTo(w)
}

func (app *application) renderPartialToString(nameFile, nameTmpl string, clients []user) string {
	// Retrieve the appropriate template set from the cache based on the page name
	// (like 'home.page.tmpl'). If no entry exists in the cache with the
	// provided name, call the serverError helper method that we made earlier.
	ts, err := template.ParseFiles(nameFile)
	if err != nil {
		fmt.Errorf("Error when parsing template : %s", nameFile)
	}
	// Initialize a new buffer.
	buff := new(bytes.Buffer)
	err = ts.ExecuteTemplate(buff, nameTmpl, clients)
	if err != nil {
		fmt.Println(err)
	}

	return buff.String()
}
func newTemplateCache(dir string) (map[string]*template.Template, error) {
	// Initialize a new map to act as the cache.
	cache := map[string]*template.Template{}

	// Use the filepath.Glob function to get a slice of all filepaths with
	// the extension '.page.tmpl'. This essentially gives us a slice of all the
	// 'page' templates for the application.
	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))

	if err != nil {

		return nil, err
	}

	// Loop through the pages one-by-one.
	for _, page := range pages {

		// Extract the file name (like 'home.page.tmpl') from the full file path
		// and assign it to the name variable.
		name := filepath.Base(page)

		// The template.FuncMap must be registered with the template set before you
		// call the ParseFiles() method. This means we have to use template.New() to
		// create an empty template set, use the Funcs() method to register the
		// template.FuncMap, and then parse the file as normal.
		ts, err := template.New(name).ParseFiles(page)
		if err != nil {
			return nil, err
		}
		// Use the ParseGlob method to add any 'layout' templates to the
		// template set (in our case, it's just the 'base' layout at the
		// moment).
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}
		// Use the ParseGlob method to add any 'partial' templates to the
		// template set (in our case, it's just the 'footer' partial at the
		// moment).
		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}
		// Add the template set to the cache, using the name of the page
		// (like 'home.page.tmpl') as the key.
		cache[name] = ts

	}
	return cache, nil
}

func (app *application) wrapDataTemplate(client *client, evenName string, template string) *dataTemplate {
	return &dataTemplate{
		CurrentUser: user{
			Pseudo: client.user.Pseudo,
			ID:     client.user.ID,
		},
		EventName: evenName,
		Template:  template,
		Users:     wrapUsersFromMap(app.ws.client),
	}
}

func wrapUsersFromMap(c map[*client]bool) (arrClient []user) {

	for client := range c {
		arrClient = append(arrClient, user{
			Pseudo: client.user.Pseudo,
			ID:     client.user.ID,
		})
	}
	return arrClient
}
