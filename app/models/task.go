package models

import (
	"go-do-the-thing/database"
	"time"
)

type Task struct {
	Id           int64                `json:"id,omitempty"`
	Name         string               `json:"name"`
	Description  string               `json:"description,omitempty"`
	AssignedTo   int64                `json:"assigned_to"`
	Status       ItemStatus           `json:"status"`
	CompleteDate *database.SqLiteTime `json:"complete_date"`
	DueDate      *database.SqLiteTime `json:"due_date"`
	CreatedBy    int64                `json:"created_by"`
	CreatedDate  *database.SqLiteTime `json:"created_date"`
	ModifiedBy   int64                `json:"modified_by"`
	ModifiedDate *database.SqLiteTime `json:"modified_date"`
	IsDeleted    bool                 `json:"is_deleted"`
	Tag          string               `json:"tag,omitempty"`
}

type ItemStatus int

const (
	Scheduled ItemStatus = iota
	Completed
)

func (t *Task) ToggleStatus(modifiedBy int64) {
	t.ModifiedBy = modifiedBy
	t.ModifiedDate = database.SqLiteNow()
	if t.Status == Scheduled {
		t.Status = Completed
		now := time.Now()
		t.CompleteDate = &database.SqLiteTime{Time: &now}
	} else {
		t.Status = Scheduled
		t.CompleteDate = &database.SqLiteTime{}
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

type TaskViewListItem struct {
	Id            int64
	Name          string
	Description   string
	AssignedTo    string
	Status        ItemStatus
	CompletedDate *database.SqLiteTime
	DueDate       *database.SqLiteTime
	Tag           string
}

func ListItemFromTask(task Task, assignedTo User) TaskViewListItem {
	return TaskViewListItem{
		Id:            task.Id,
		Name:          task.Name,
		Description:   task.Description,
		AssignedTo:    assignedTo.FullName,
		Status:        task.Status,
		CompletedDate: task.CompleteDate,
		DueDate:       task.DueDate,
		Tag:           task.Tag,
	}
}
