package form_models

type ErrorForm interface {
	GetErrors() map[string]string
	SetError(name, value string)
}
