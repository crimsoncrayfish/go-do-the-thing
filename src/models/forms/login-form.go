package form_models

type LoginForm struct {
	Email    string
	Password string
	Errors   map[string]string
}

func NewLoginForm() LoginForm {
	return LoginForm{
		Email:    "",
		Password: "",
		Errors:   make(map[string]string),
	}
}

func (f *LoginForm) GetErrors() map[string]string {
	return f.Errors
}
func (f *LoginForm) SetError(name, value string) {
	f.Errors[name] = value
}
