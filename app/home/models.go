package home

import (
	"go-do-the-thing/app/shared/models"
	"go-do-the-thing/helpers"
	"go-do-the-thing/helpers/slog"
	"go-do-the-thing/middleware"
	"net/http"
)

type Screens struct {
	NavBar models.NavBarObject
}

func SetupHome(router *http.ServeMux, templates helpers.Templates, mw_stack middleware.Middleware) {
	logger := slog.NewLogger("Home")
	logger.Info("Setting up the Home screen")
	handler := New(templates, logger)
	router.Handle("/", mw_stack(http.HandlerFunc(handler.Index)))
	router.Handle("GET /home", mw_stack(http.HandlerFunc(handler.Home)))
}
