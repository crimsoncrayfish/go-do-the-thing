package handlers

import (
	"net/http"
)

func AcceptHeaderSwitch(w http.ResponseWriter, r *http.Request, jsonFunc func(w http.ResponseWriter, r *http.Request), uiFunc func(w http.ResponseWriter, r *http.Request)) {
	contentType := r.Header.Get("accept")
	switch contentType {
	case "application/json":
		w.Header().Set("Content-Type", "application/json")
		jsonFunc(w, r)
	case "text/html":
		w.Header().Set("Content-Type", "text/html")
		uiFunc(w, r)
	default:
		http.Error(w, "No Content-type specified", http.StatusInternalServerError)
	}
}

func Redirect(location string, w http.ResponseWriter) {
	w.Header().Set("HX-Location", location)
}

func RedirectFull(location string, w http.ResponseWriter) {
	w.Header().Set("HX-Redirect", location)
}
