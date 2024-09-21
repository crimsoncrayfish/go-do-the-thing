package models

import (
	"go-do-the-thing/src/database"
	"time"
)

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
func NewDefaultTaskForm() TaskForm {
	duedate := time.Now().Add(time.Duration(time.Hour * 24))
	return TaskForm{
		Task: TaskView{
			DueDate: &database.SqLiteTime{Time: &duedate},
		},
		Errors: make(map[string]string),
	}
}
