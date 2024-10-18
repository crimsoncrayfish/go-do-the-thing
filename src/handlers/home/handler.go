package home

import (
	home_templ "go-do-the-thing/src/handlers/home/templ"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	"go-do-the-thing/src/models"
	"net/http"
)

type HomeHandler struct {
	logger slog.Logger
}

var activeScreens Screens

func SetupHomeHandler(router *http.ServeMux, mw_stack middleware.Middleware) {
	logger := slog.NewLogger("Home")
	logger.Info("Setting up the Home screen")
	activeScreens = Screens{
		models.NavBarObject{
			ActiveScreens: models.ActiveScreens{IsHome: true},
		},
	}
	handler := &HomeHandler{
		logger: logger,
	}
	router.Handle("/", mw_stack(http.HandlerFunc(handler.Index)))
	router.Handle("GET /home", mw_stack(http.HandlerFunc(handler.Home)))
}

type Screens struct {
	NavBar models.NavBarObject
}

func (h *HomeHandler) Index(w http.ResponseWriter, r *http.Request) {
	// Get currentUser details
	_, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, h.logger, "user auth failed unsuccessfully")

	if err := home_templ.Index(activeScreens.NavBar).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "Failed to execute template for the home page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *HomeHandler) Home(w http.ResponseWriter, r *http.Request) {
	// Get currentUser details
	_, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, h.logger, "user auth failed unsuccessfully")

	if err := home_templ.Index(activeScreens.NavBar).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "Failed to execute template for the home page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
