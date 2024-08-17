package helpers

import (
	"errors"
	"go-do-the-thing/app/shared"
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
		shared.HttpError("No Content-type specified", errors.New("no content-type specified in request"), w)
	}
}
