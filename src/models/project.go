package models

import (
	"time"
)

type Project struct {
	Id           int
	Name         string
	Description  string
	Owner        int
	StartDate    time.Time
	DueDate      time.Time
	CreatedBy    int
	CreatedDate  time.Time
	ModifiedBy   int
	ModifiedDate time.Time
	IsComplete   bool
	IsDeleted    bool
}

type ProjectTag struct {
	ProjectId int
	TagId     int
}

type ProjectUser struct {
	ProjectId int
	UserId    int
	RoleId    RoleEnum
}
