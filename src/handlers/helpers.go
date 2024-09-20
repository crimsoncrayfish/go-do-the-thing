package handlers

import (
	"errors"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/models"
	"net/http"
)

func AcceptHeaderSwitch(w http.ResponseWriter, r *http.Request, jsonFunc func(w http.ResponseWriter, r *http.Request), uiFunc func(w http.ResponseWriter, r *http.Request)) {
	contentType := r.Header.Get("accept")
	if contentType == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		jsonFunc(w, r)
	} else if contentType == "text/html" {
		w.Header().Set("Content-Type", "text/html")
		uiFunc(w, r)
	} else {
		HttpError("No Content-type specified", errors.New("no content-type specified in request"), w)
	}
}

type ErrorPage struct {
	NavBar       models.NavBarObject
	Message      string
	ErrorMessage string
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
	http.Error(w, message, http.StatusInternalServerError)
}

func Redirect(location string, w http.ResponseWriter) {
	// TODO: Add ability to let user know why redirect happened (message on screen?)
	w.Header().Set("HX-Location", location)
}
