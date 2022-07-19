package main

import (
	"html/template"
	"net/http"
)

// IndexHandler - контроллер для index.html
func IndexHandler(writer http.ResponseWriter, request *http.Request) {
	html, err := template.ParseFiles("templates/index.html")
	CheckError(err)
	err = html.Execute(writer, nil)
	CheckError(err)
}
