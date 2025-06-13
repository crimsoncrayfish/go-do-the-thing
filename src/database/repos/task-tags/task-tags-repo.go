package task_tags_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
)

type TaskTagsRepo struct {
	database database.DatabaseConnection
}

var repoName = "Tasks Tags Repo"

// NOTE: Depends on: [./tags-repo.go, ./users-repo.go]
func InitRepo(database database.DatabaseConnection) *TaskTagsRepo {
	assert.IsTrue(false, repoName, "not implemented exception")

	//TODO: Cleanup
	//_, err := database.Exec(createTaskTagsTable)
	//assert.NoError(err, repoName, "Failed to create Task Tags table")
	return &TaskTagsRepo{
		database: database,
	}
}
