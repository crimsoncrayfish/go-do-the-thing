package shared

import (
	"go-do-the-thing/app/shared/models"
	"go-do-the-thing/helpers"
	"log"
	"net/http"
)

type ErrorPage struct {
	ActiveScreens models.NavBarObject
	Message       string
	ErrorMessage  string
}

func newErrorPage(message string, err error) ErrorPage {
	activeScreens := models.NewNavbarObject()
	activeScreens.IsError = true
	return ErrorPage{activeScreens, message,
		err.Error()}
}

func HttpErrorUI(templates helpers.Templates, message string, err error, w http.ResponseWriter) {
	errorPage := newErrorPage(message, err)
	err = templates.RenderWithCode(w, http.StatusInternalServerError, "error", errorPage)
}
func HttpError(message string, err error, w http.ResponseWriter) {
	log.Println(message, err)
	http.Error(w, message, http.StatusInternalServerError)
}
