package models

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/assert"
	"time"
)

type Project struct {
	Id           int64
	Name         string
	Description  string
	Owner        int64
	StartDate    *time.Time
	DueDate      *time.Time
	CreatedBy    int64
	CreatedDate  *time.Time
	ModifiedBy   int64
	ModifiedDate *time.Time
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
	Id           int64
	Name         string
	Description  string
	Owner        UserView
	StartDate    *time.Time
	DueDate      *time.Time
	CreatedBy    UserView
	CreatedDate  *time.Time
	ModifiedBy   UserView
	ModifiedDate *time.Time
	IsComplete   bool
	IsDeleted    bool
}

func (p *Project) ToViewModel(owner, createdBy, modifiedBy *User) ProjectView {
	assert.NotNil(p, helpers.PrevCallerName(2), "project cant be nil")
	assert.NotNil(owner, helpers.PrevCallerName(2), "project owner cant be nil")
	assert.NotNil(createdBy, helpers.PrevCallerName(2), "project creator cant be nil")
	assert.NotNil(modifiedBy, helpers.PrevCallerName(2), "project modifier cant be nil")

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

func ProjectListToMap(projects []ProjectView) map[int64]string {
	new_map := make(map[int64]string, len(projects))
	for _, project := range projects {
		new_map[project.Id] = project.Name
	}
	return new_map
}
