package form_models

import (
	"go-do-the-thing/src/models"
	"time"
)

type TaskForm struct {
	Task   models.TaskView
	Errors map[string]string
}

func NewTaskForm() TaskForm {
	return TaskForm{
		Task:   models.TaskView{},
		Errors: make(map[string]string),
	}
}

func NewDefaultTaskForm() TaskForm {
	duedate := time.Now().Add(time.Duration(time.Hour * 24))
	return TaskForm{
		Task: models.TaskView{
			DueDate: &duedate,
		},
		Errors: make(map[string]string),
	}
}

func (f *TaskForm) GetErrors() map[string]string {
	return f.Errors
}

func (f *TaskForm) SetError(name, value string) {
	f.Errors[name] = value
}

func (f *TaskForm) SetProject(project_id int64) {
	f.Task.ProjectId = project_id
}
