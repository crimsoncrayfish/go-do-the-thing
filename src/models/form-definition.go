package models

import (
	"fmt"
	"net/http"
)

type FormData struct {
	Values map[string]string
	Errors map[string]string
	Submit string
}

func NewFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
		Submit: "Create",
	}
}

func GetRequiredPropertyFromRequest(r *http.Request, propName, title string) (string, error) {
	value := r.FormValue(propName)
	if len(value) == 0 {
		return value, fmt.Errorf("%s is required", title)
	}

	return value, nil
}

func GetPropertyFromRequest(r *http.Request, propName, title string) string {
	return r.FormValue(propName)
}
