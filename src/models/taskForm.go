package models

type TaskForm struct {
	Task   TaskView
	Errors map[string]string
}

func NewTaskForm() TaskForm {
	return TaskForm{
		Task:   TaskView{},
		Errors: make(map[string]string),
	}
}
