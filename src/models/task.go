package models

import (
	"go-do-the-thing/src/database"
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
	Project      int64                `json:"project_id"`
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
		t.CompleteDate = database.SqLiteNow()
	} else {
		t.Status = Scheduled
		t.CompleteDate = &database.SqLiteTime{}
	}
}

func (t *Task) IsValid() (bool, map[string]string) {
	errs := make(map[string]string)
	isValid := true

	if t.DueDate.Before(database.SqLiteNow()) {
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
	CompletedDate *database.SqLiteTime
	DueDate       *database.SqLiteTime
	CreatedDate   *database.SqLiteTime
	CreatedBy     string
	Project       int64
}

func (t *Task) ToViewModel(assignedTo, createdBy User) TaskView {
	return TaskView{
		Id:            t.Id,
		Name:          t.Name,
		Description:   t.Description,
		AssignedTo:    assignedTo.FullName,
		Status:        t.Status,
		CompletedDate: t.CompleteDate,
		CreatedDate:   t.CreatedDate,
		CreatedBy:     createdBy.FullName,
		DueDate:       t.DueDate,
		Project:       t.Project,
	}
}
