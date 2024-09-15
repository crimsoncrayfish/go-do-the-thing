package models

import "net/http"

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

func GetRequiredPropertyFromRequest(r *http.Request, propName string, formData FormData, keepValue bool) (string, FormData) {
	value := r.FormValue(propName)
	if len(value) == 0 {
		formData.Errors[propName] = propName + " is required"
		return value, formData
	}
	if keepValue {
		formData.Values[propName] = value
	}

	return value, formData
}

func GetOptionalPropertyFromRequest(r *http.Request, propName string, formData FormData, keepValue bool) (string, FormData) {
	value := r.FormValue(propName)
	if keepValue {
		formData.Values[propName] = value
	}
	return value, formData
}
