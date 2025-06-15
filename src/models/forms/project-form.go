package form_models

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/models"
	"time"
)

type ProjectForm struct {
	Project models.ProjectView
	Errors  map[string]string
}

func NewProjectForm() ProjectForm {
	return ProjectForm{
		Project: models.ProjectView{},
		Errors:  make(map[string]string),
	}
}

func NewDefaultProjectForm() ProjectForm {
	duedate := time.Now().Add(time.Duration(time.Hour * 24))
	return ProjectForm{
		Project: models.ProjectView{
			DueDate: &duedate,
		},
		Errors: make(map[string]string),
	}
}

func (f *ProjectForm) GetErrors() map[string]string {
	return f.Errors
}
func (f *ProjectForm) SetError(name, value string) {
	f.Errors[name] = value
}
