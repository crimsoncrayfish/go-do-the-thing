package user_models

import (
	"go-do-the-thing/app/models"
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
	AssignedTo   int64                `json:"assigned_to"`
	DueDate      *database.SqLiteTime `json:"due_date"`
	CreatedBy    int64                `json:"created_by"`
	DateCreated  *database.SqLiteTime `json:"create_date"`
	UpdatedBy    int64                `json:"updated_by"`
	DateUpdated  *database.SqLiteTime `json:"update_date"`
	IsDeleted    bool                 `json:"is_deleted"`
	Tag          string               `json:"tag,omitempty"`
}

func (t *Task) ToViewModel(assignedTo models.User) TaskView {
	return TaskView{
		Name:         t.Name,
		Description:  t.Description,
		Status:       t.Status,
		CompleteDate: *t.CompleteDate.Time,
		AssignedTo:   assignedTo.FullName,
		DueDate:      *t.DueDate.Time,
		Tag:          t.Tag,
	}
}

type ItemStatus int

const (
	TaskScheduled ItemStatus = iota
	TaskCompleted
)

func (t *Task) ToggleStatus() {
	if t.Status == TaskScheduled {
		t.Status = TaskCompleted
		now := time.Now()
		t.CompleteDate = &database.SqLiteTime{Time: &now}
	} else {
		t.Status = TaskScheduled
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

func (t *Task) FormDataFromItemNoValidation(assignedTo string) models.FormData {
	formData := models.NewFormData()
	formData.Values["name"] = t.Name
	formData.Values["description"] = t.Description
	formData.Values["assigned_to"] = assignedTo
	formData.Values["due_date"] = t.DueDate.StringF(constants.DateFormat)
	formData.Values["tag"] = t.Tag

	return formData
}

func (t *Task) FormDataFromItem(assignedTo string) (models.FormData, bool) {
	formData := t.FormDataFromItemNoValidation(assignedTo)
	isValid, errs := t.IsValid()
	if !isValid {
		formData.Errors = errs
	}
	return formData, isValid
}
