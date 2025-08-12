package task_tags_repo

import (
	"go-do-the-thing/src/database"
)

type TaskTagsRepo struct {
	database database.DatabaseConnection
}

var repoName = "Tasks Tags Repo"

// NOTE: Depends on: [./tags_repo.go, ./users_repo.go]
func InitRepo(database database.DatabaseConnection) *TaskTagsRepo {
	return &TaskTagsRepo{
		database: database,
	}
}
