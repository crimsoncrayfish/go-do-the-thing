package user_models

import (
	"time"
)

type TaskView struct {
	Name         string
	Description  string
	Status       ItemStatus
	CompleteDate time.Time
	AssignedTo   string
	DueDate      time.Time
	Tag          string
}
