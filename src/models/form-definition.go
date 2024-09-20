package models

import (
	"errors"
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

func GetPropertyFromRequest(r *http.Request, propName string, required bool) (string, error) {
	value := r.FormValue(propName)
	if len(value) == 0 && required {
		return value, errors.New(fmt.Sprintf("Failed to get value for '%s'", propName))
	}

	return value, nil
}
