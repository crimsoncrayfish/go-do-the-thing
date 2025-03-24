package models

import (
	"go-do-the-thing/src/database"
)

type Project struct {
	Id           int
	Name         string
	Description  string
	Owner        int64
	StartDate    *database.SqLiteTime
	DueDate      *database.SqLiteTime
	CreatedBy    int64
	CreatedDate  *database.SqLiteTime
	ModifiedBy   int64
	ModifiedDate *database.SqLiteTime
	IsComplete   bool
	IsDeleted    bool
}

type ProjectView struct {
	Id           int
	Name         string
	Description  string
	Owner        UserView
	StartDate    *database.SqLiteTime
	DueDate      *database.SqLiteTime
	CreatedBy    UserView
	CreatedDate  *database.SqLiteTime
	ModifiedBy   UserView
	ModifiedDate *database.SqLiteTime
	IsComplete   bool
	IsDeleted    bool
}

func ProjectToViewModel(project Project, createdBy User) ProjectView {
	return ProjectView{
		Id:          project.Id,
		Name:        project.Name,
		Description: project.Description,
		CreatedDate: project.CreatedDate,
		CreatedBy:   UserToViewModel(createdBy),
		DueDate:     project.DueDate,
	}
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
