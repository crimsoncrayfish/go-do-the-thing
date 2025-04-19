package models

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
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

func (p *Project) AssertHealthyNew() {
	source := "Model.Project"

	assert.NotNil(p, source, "project model struct not healthy - project is nil")

	assert.NotEqual(p.Name, "", source, "Name")
	assert.NotEqual(p.Description, "", source, "Description")

	assert.NotNil(p.StartDate, source, "project model struct not healthy, nil StartDate")
	assert.NotNil(p.DueDate, source, "project model struct not healthy, nil DueDate")

	assert.NotEqual(p.CreatedBy, 0, source, "CreatedBy")
	assert.NotNil(p.CreatedDate, source, "project model struct not healthy, nil CreatedDate")
	assert.NotEqual(p.ModifiedBy, 0, source, "ModifiedBy")
	assert.NotNil(p.ModifiedDate, source, "project model struct not healthy, nil ModifiedDate")
}

func (p *Project) AssertHealthy() {
	source := "Model.Project"
	assert.NotEqual(p.Id, 0, source, "Id")
	p.AssertHealthyNew()
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

func (p *Project) ToViewModel(owner, createdBy, modifiedBy *User) ProjectView {
	return ProjectView{
		Id:           p.Id,
		Name:         p.Name,
		Description:  p.Description,
		Owner:        owner.ToViewModel(),
		CreatedDate:  p.CreatedDate,
		CreatedBy:    createdBy.ToViewModel(),
		ModifiedDate: p.ModifiedDate,
		ModifiedBy:   modifiedBy.ToViewModel(),
		DueDate:      p.DueDate,
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
