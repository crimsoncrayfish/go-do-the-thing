package construction

import (
	"go-do-the-thing/src/helpers"
	app_errors "go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	"go-do-the-thing/src/models"
	"net/http"

	templ_construction "go-do-the-thing/src/handlers/construction/templ"
)

type ConstructionHandler struct {
	logger slog.Logger
}

var source = "ConstructionHandler"

func SetupConstructionHandler(router *http.ServeMux, mw_stack middleware.Middleware) {
	logger := slog.NewLogger(source)
	logger.Info("Setting up the Construction Handler")
	handler := &ConstructionHandler{
		logger: logger,
	}
	router.Handle("GET /habbits", mw_stack(http.HandlerFunc(handler.Habbits)))
}

func (h *ConstructionHandler) Habbits(w http.ResponseWriter, r *http.Request) {
	_, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		app_errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	// Check if this is an HTMX request by looking for HX-Request header
	if r.Header.Get("HX-Request") == "true" {
		// HTMX request - return just the page content
		if err := templ_construction.UnderConstructionPage(models.ScreenHabbits).Render(r.Context(), w); err != nil {
			app_errors.InternalServerError(w, r, h.logger, err, "Failed to display under construction page content")
			return
		}
	} else {
		// Direct page load - return full page with body wrapper
		if err := templ_construction.UnderConstructionWithBody(models.ScreenHabbits).Render(r.Context(), w); err != nil {
			app_errors.InternalServerError(w, r, h.logger, err, "Failed to display under construction page")
			return
		}
	}
}
