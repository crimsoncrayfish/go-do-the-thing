package form_models

type RegistrationForm struct {
	Name   string
	Email  string
	Errors map[string]string
}

func NewRegistrationForm() RegistrationForm {
	return RegistrationForm{
		Name:  "",
		Email: "",

		Errors: make(map[string]string),
	}
}

func (f *RegistrationForm) GetErrors() map[string]string {
	return f.Errors
}
func (f *RegistrationForm) SetError(name, value string) {
	f.Errors[name] = value
}
