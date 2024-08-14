package home

import (
	"fmt"
	"go-do-the-thing/app/shared/models"
	"go-do-the-thing/helpers"
	"net/http"
)

type Screens struct {
	ActiveScreens models.NavBarObject
}

func SetupHome(router *http.ServeMux, templates helpers.Templates) {
	fmt.Println("Setting up the Home screen")
	handler := New(templates)
	router.HandleFunc("/", handler.Index)
	router.HandleFunc("GET /home", handler.Home)
}
