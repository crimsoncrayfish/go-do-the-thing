package models

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
