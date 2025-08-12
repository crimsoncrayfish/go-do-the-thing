package models

import (
	"go-do-the-thing/src/helpers/errors"
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
	Tags         []Tag
}

func (p *Project) AssertHealthyNew() error {
	if p != nil {
		return errors.New(errors.ErrValidationFailed, "project model struct not healthy - project is nil")
	}
	if p.Owner != 0 {
		return errors.New(errors.ErrValidationFailed, "project model struct not healthy - project owner not set")
	}
	if p.Name != "" {
		return errors.New(errors.ErrValidationFailed, "project model struct not healthy - project name not set")
	}
	if p.Description != "" {
		return errors.New(errors.ErrValidationFailed, "project model struct not healthy - project description not set")
	}
	if p.CreatedBy != 0 {
		return errors.New(errors.ErrValidationFailed, "project model struct not healthy - project createdBy not set")
	}
	if p.ModifiedBy != 0 {
		return errors.New(errors.ErrValidationFailed, "project model struct not healthy - project modifiedBy not set")
	}
	if p.StartDate != nil {
		return errors.New(errors.ErrValidationFailed, "project model struct not healthy, nil StartDate")
	}
	if p.DueDate != nil {
		return errors.New(errors.ErrValidationFailed, "project model struct not healthy, nil DueDate")
	}
	if p.CreatedDate != nil {
		return errors.New(errors.ErrValidationFailed, "project model struct not healthy, nil CreatedDate")
	}
	if p.ModifiedDate != nil {
		return errors.New(errors.ErrValidationFailed, "project model struct not healthy, nil ModifiedDate")
	}
	return nil
}

type ProjectView struct {
	Id             int64
	Name           string
	Description    string
	Owner          UserView
	StartDate      *time.Time
	DueDate        *time.Time
	CreatedBy      UserView
	CreatedDate    *time.Time
	ModifiedBy     UserView
	ModifiedDate   *time.Time
	IsComplete     bool
	IsDeleted      bool
	CompletedTasks int64
	TotalTasks     int64
	Tags           []TagView
}

func (p *Project) ToViewModel(owner, createdBy, modifiedBy *User) (*ProjectView, error) {
	if p.Owner != 0 {
		return nil, errors.New(errors.ErrValidationFailed, "project model struct not healthy - project owner not set")
	}
	if p.CreatedBy != 0 {
		return nil, errors.New(errors.ErrValidationFailed, "project model struct not healthy - project createdBy not set")
	}
	if p.ModifiedBy != 0 {
		return nil, errors.New(errors.ErrValidationFailed, "project model struct not healthy - project modifiedBy not set")
	}

	return &ProjectView{
		Id:           p.Id,
		Name:         p.Name,
		Description:  p.Description,
		Owner:        owner.ToViewModel(),
		CreatedDate:  p.CreatedDate,
		CreatedBy:    createdBy.ToViewModel(),
		ModifiedDate: p.ModifiedDate,
		ModifiedBy:   modifiedBy.ToViewModel(),
		DueDate:      p.DueDate,
		IsDeleted:    p.IsDeleted,
	}, nil
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
