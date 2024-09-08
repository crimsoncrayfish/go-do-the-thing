package home

import (
	"fmt"
	"go-do-the-thing/app/shared/models"
	"go-do-the-thing/helpers"
	"go-do-the-thing/middleware"
	"net/http"
)

type Screens struct {
	ActiveScreens models.NavBarObject
}

func SetupHome(router *http.ServeMux, templates helpers.Templates, mw_stack middleware.Middleware) {
	fmt.Println("Setting up the Home screen")
	handler := New(templates)
	router.Handle("/", mw_stack(http.HandlerFunc(handler.Index)))
	router.Handle("GET /home", mw_stack(http.HandlerFunc(handler.Home)))
}
