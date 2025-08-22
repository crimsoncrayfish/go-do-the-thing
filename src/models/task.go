// Package models implements the models for this project
package models

import (
	"go-do-the-thing/src/helpers/errors"
	"time"
)

type Task struct {
	ID           int64      `json:"id,omitempty"`
	Name         string     `json:"name"`
	Description  string     `json:"description,omitempty"`
	AssignedTo   int64      `json:"assigned_to"`
	Status       ItemStatus `json:"status"`
	CompleteDate *time.Time `json:"complete_date"`
	DueDate      *time.Time `json:"due_date"`
	CreatedBy    int64      `json:"created_by"`
	CreatedDate  *time.Time `json:"created_date"`
	ModifiedBy   int64      `json:"modified_by"`
	ModifiedDate *time.Time `json:"modified_date"`
	IsDeleted    bool       `json:"is_deleted"`
	Project      int64      `json:"project_id"`
	Tags         []Tag
}

type ItemStatus int

const (
	Scheduled ItemStatus = iota
	Completed
)

func (t *Task) ToggleStatus(modifiedBy int64) {
	now := time.Now()
	t.ModifiedBy = modifiedBy
	t.ModifiedDate = &now
	if t.Status == Scheduled {
		t.Status = Completed
		t.CompleteDate = &now
	} else {
		t.Status = Scheduled
		t.CompleteDate = &time.Time{}
	}
}

func (t *Task) IsValid() (bool, map[string]string) {
	errs := make(map[string]string)
	isValid := true

	if t.DueDate.Before(time.Now()) {
		isValid = false
		errs["due_date"] = "Due date is before now"
	}
	return isValid, errs
}

type TaskView struct {
	ID            int64
	Name          string
	Description   string
	AssignedTo    UserView
	Status        ItemStatus
	CompletedDate *time.Time
	DueDate       *time.Time
	CreatedDate   *time.Time
	CreatedBy     UserView
	ModifiedDate  *time.Time
	ModifiedBy    UserView
	ProjectID     int64
	ProjectName   string
	InProgress    bool
	TimeSpent     time.Duration
	IsDeleted     bool
	Tags          []TagView
}

func (t *TaskView) IsOverDue() bool {
	if t.DueDate == nil {
		return false
	}
	if t.DueDate.Before(time.Now().Add(-24 * time.Hour)) {
		return true
	}
	return false
}

func (t *Task) ToViewModel(assignedTo, createdBy, modifiedBy *User, project Project) (*TaskView, error) {
	if assignedTo == nil {
		return nil, errors.New(errors.ErrValidationFailed, "task assigned to cant be nil")
	}
	if createdBy == nil {
		return nil, errors.New(errors.ErrValidationFailed, "task created by cant be nil")
	}
	if modifiedBy == nil {
		return nil, errors.New(errors.ErrValidationFailed, "task modified by cant be nil")
	}
	tagViews := make([]TagView, len(t.Tags))
	for i, tag := range t.Tags {
		tagViews[i] = *tag.ToView()
	}
	return &TaskView{
		ID:            t.ID,
		Name:          t.Name,
		Description:   t.Description,
		AssignedTo:    assignedTo.ToViewModel(),
		Status:        t.Status,
		CompletedDate: t.CompleteDate,
		CreatedDate:   t.CreatedDate,
		CreatedBy:     createdBy.ToViewModel(),
		ModifiedDate:  t.ModifiedDate,
		ModifiedBy:    modifiedBy.ToViewModel(),
		DueDate:       t.DueDate,
		ProjectID:     t.Project,
		ProjectName:   project.Name,
		IsDeleted:     t.IsDeleted,
		Tags:          tagViews,
	}, nil
}
