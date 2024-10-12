package models

import "go-do-the-thing/src/database"

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
	IsDeleted    bool
}
