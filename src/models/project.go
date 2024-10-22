package models

import (
	"go-do-the-thing/src/database"
)

type Project struct {
	Id           int
	Name         string
	Description  string
	Owner        int
	StartDate    *database.SqLiteTime
	DueDate      *database.SqLiteTime
	CreatedBy    int
	CreatedDate  *database.SqLiteTime
	ModifiedBy   int
	ModifiedDate *database.SqLiteTime
	IsComplete   bool
	IsDeleted    bool
}

type ProjectUI struct {
	Id           int
	Name         string
	Description  string
	Owner        User
	StartDate    *database.SqLiteTime
	DueDate      *database.SqLiteTime
	CreatedBy    User
	CreatedDate  *database.SqLiteTime
	ModifiedBy   User
	ModifiedDate *database.SqLiteTime
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
