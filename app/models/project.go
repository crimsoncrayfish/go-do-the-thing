package models

import "go-do-the-thing/database"

type Project struct {
	Id          int
	Name        string
	Description string
	Status      ProjectStatus
	CreatedBy   int
	CreateDate  database.SqLiteTime
	IsDeleted   bool
}

type ProjectStatus int

const (
	ProjectBacklog ProjectStatus = iota
	ProjectInProgress
	ProjectClosed
)
