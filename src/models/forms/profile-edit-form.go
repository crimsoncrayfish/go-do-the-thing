package form_models

import "go-do-the-thing/src/models"

type ProfileEditForm struct {
	Email  string
	Name   string
	Errors map[string]string
}

func NewProfileEditForm(user models.UserView) ProfileEditForm {
	return ProfileEditForm{
		Email:  "",
		Name:   "",
		Errors: make(map[string]string),
	}
}

func (f *ProfileEditForm) GetErrors() map[string]string {
	return f.Errors
}

func (f *ProfileEditForm) SetError(name, value string) {
	f.Errors[name] = value
}
