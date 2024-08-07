package home

import (
	"fmt"
	"go-do-the-thing/helpers"
	"go-do-the-thing/navigation"
	"net/http"
)

type Screens struct {
	ActiveScreens navigation.NavBarObject
}

func SetupHome(router *http.ServeMux, templates helpers.Templates) {
	fmt.Println("Setting up the Home screen")
	handler := New(templates)
	router.HandleFunc("GET /home", handler.Index)
}
