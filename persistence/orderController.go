package main

import (
	"html/template"
	"net/http"
)

// OrderHandler Контроллер для страницы order.html
func OrderHandler(writer http.ResponseWriter, request *http.Request) {
	uid := request.FormValue("uid")
	orderm, ok := GetByUID(uid)
	if ok {
		html, err := template.ParseFiles("templates/order.html")
		CheckError(err)
		err = html.Execute(writer, orderm)
		CheckError(err)
	} else {
		html, err := template.ParseFiles("templates/error.html")
		CheckError(err)
		err = html.Execute(writer, nil)
		CheckError(err)
	}
}
