package models

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/assert"
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
	AssignedTo    UserView
	Status        ItemStatus
	CompletedDate *database.SqLiteTime
	DueDate       *database.SqLiteTime
	CreatedDate   *database.SqLiteTime
	CreatedBy     UserView
	ModifiedDate  *database.SqLiteTime
	ModifiedBy    UserView
	ProjectId     int64
	ProjectName   string
	InProgress    bool
	TimeSpent     time.Duration
	IsDeleted     bool
}

func (t *Task) ToViewModel(assignedTo, createdBy, modifiedBy *User, project Project) *TaskView {
	assert.NotNil(assignedTo, helpers.PrevCallerName(2), "task assigned to cant be nil")
	assert.NotNil(createdBy, helpers.PrevCallerName(2), "task creator cant be nil")
	assert.NotNil(modifiedBy, helpers.PrevCallerName(2), "task modifier cant be nil")
	return &TaskView{
		Id:            t.Id,
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
		ProjectId:     t.Project,
		ProjectName:   project.Name,
		IsDeleted:     t.IsDeleted,
	}
}
