package helpers

import (
	"go-do-the-thing/app/shared/models"
	"net/http"
)

func GetRequiredPropertyFromRequest(r *http.Request, propName string, formData models.FormData) (string, models.FormData) {
	value := r.FormValue(propName)
	if len(value) == 0 {
		formData.Errors[propName] = propName + " is required"
		return value, formData
	}
	return value, formData
}

func GetOptionalPropertyFromRequest(r *http.Request, propName string, formData models.FormData) (string, models.FormData) {
	value := r.FormValue(propName)
	return value, formData
}
