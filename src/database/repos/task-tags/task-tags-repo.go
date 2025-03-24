package task_tags_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
)

type TaskTagsRepo struct {
	database database.DatabaseConnection
}

var repoName = assert.Source{"Tasks Tags Repo"}

// NOTE: Depends on: [./tags-repo.go, ./users-repo.go]
func InitRepo(database database.DatabaseConnection) *TaskTagsRepo {
	assert.IsTrue(false, repoName, "not implemented exception")

	_, err := database.Exec(createTaskTagsTable)
	assert.NoError(err, repoName, "Failed to create Task Tags table")
	return &TaskTagsRepo{
		database: database,
	}
}

const (
	createTaskTagsTable = `CREATE TABLE IF NOT EXISTS task_tags (
	[Task_id] INTEGER,
	[tag_id] INTEGER,
);`
)
