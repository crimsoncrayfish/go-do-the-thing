package home

import (
	"fmt"
	"go-do-the-thing/helpers"
	"net/http"
)

type Count struct {
	Count int
}

type Contact struct {
	Name  string
	Email string
}

type HomeObject struct {
	Count    int
	Contacts []Contact
}

func SetupHome(router *http.ServeMux, templates helpers.Templates) {
	fmt.Println("Setting up the Home screen")
	handler := New(templates)
	router.HandleFunc("GET /home", handler.Index)

	router.HandleFunc("POST /increment", handler.IncrementCount)
	router.HandleFunc("POST /contacts/add", handler.AddContact)
}
