package project

import (
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	"net/http"
)

type ProjectHandler struct {
	logger slog.Logger
}

func SetupProjectHandler(router *http.ServeMux, mw_stack middleware.Middleware) {

}
