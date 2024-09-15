package models

import (
	"go-do-the-thing/database"
	"go-do-the-thing/helpers/constants"
	"time"
)

type Task struct {
	Id           int64                `json:"id,omitempty"`
	Name         string               `json:"name"`
	Description  string               `json:"description,omitempty"`
	Status       ItemStatus           `json:"status"`
	CompleteDate *database.SqLiteTime `json:"complete_date"`
	AssignedTo   string               `json:"assigned_to"`
	DueDate      *database.SqLiteTime `json:"due_date"`
	CreatedBy    string               `json:"created_by"`
	CreateDate   *database.SqLiteTime `json:"create_date"`
	IsDeleted    bool                 `json:"is_deleted"`
	Tag          string               `json:"tag,omitempty"`
	AssignedUser int64                `json:"assigned_to_user,omitempty"`
}

type ItemStatus int

const (
	Scheduled ItemStatus = iota
	Completed
)

func (t *Task) ToggleStatus() {
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

func (t *Task) FormDataFromItemNoValidation() FormData {
	formData := NewFormData()
	formData.Values["name"] = t.Name
	formData.Values["description"] = t.Description
	formData.Values["assigned_to"] = t.AssignedTo
	formData.Values["due_date"] = t.DueDate.StringF(constants.DateFormat)
	formData.Values["tag"] = t.Tag

	return formData
}

func (t *Task) FormDataFromItem() (FormData, bool) {
	formData := t.FormDataFromItemNoValidation()
	isValid, errs := t.IsValid()
	if !isValid {
		formData.Errors = errs
	}
	return formData, isValid
}
