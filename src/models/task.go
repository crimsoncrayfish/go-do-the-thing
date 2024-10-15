package models

import (
	"time"
)

type Task struct {
	Id           int64      `json:"id,omitempty"`
	Name         string     `json:"name"`
	Description  string     `json:"description,omitempty"`
	AssignedTo   int64      `json:"assigned_to"`
	Status       ItemStatus `json:"status"`
	CompleteDate time.Time  `json:"complete_date"`
	DueDate      time.Time  `json:"due_date"`
	CreatedBy    int64      `json:"created_by"`
	CreatedDate  time.Time  `json:"created_date"`
	ModifiedBy   int64      `json:"modified_by"`
	ModifiedDate time.Time  `json:"modified_date"`
	IsDeleted    bool       `json:"is_deleted"`
	Project      int64      `json:"project_id"`
}

type ItemStatus int

const (
	Scheduled ItemStatus = iota
	Completed
)

func (t *Task) ToggleStatus(modifiedBy int64) {
	t.ModifiedBy = modifiedBy
	t.ModifiedDate = time.Now()
	if t.Status == Scheduled {
		t.Status = Completed
		t.CompleteDate = time.Now()
	} else {
		t.Status = Scheduled
		t.CompleteDate = time.Time{}
	}
}

func (t *Task) IsValid() (bool, map[string]string) {
	errs := make(map[string]string)
	isValid := true

	now := time.Now()
	if t.DueDate.Before(now) {
		isValid = false
		errs["due_date"] = "Due date is before now"
	}
	return isValid, errs
}

type TaskView struct {
	Id            int64
	Name          string
	Description   string
	AssignedTo    string
	Status        ItemStatus
	CompletedDate time.Time
	DueDate       time.Time
	CreatedDate   time.Time
	CreatedBy     string
	Project       int64
}

func TaskToViewModel(task Task, assignedTo, createdBy User) TaskView {
	return TaskView{
		Id:            task.Id,
		Name:          task.Name,
		Description:   task.Description,
		AssignedTo:    assignedTo.FullName,
		Status:        task.Status,
		CompletedDate: task.CompleteDate,
		CreatedDate:   task.CreatedDate,
		CreatedBy:     createdBy.FullName,
		DueDate:       task.DueDate,
		Project:       task.Project,
	}
}
