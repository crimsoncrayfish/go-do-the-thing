package repos

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
)

type TaskTagsRepo struct {
	database database.DatabaseConnection
}

const TaskTagsRepoName = "Task-tags"

func initTaskTagsRepo(database database.DatabaseConnection) *TaskTagsRepo {
	logger := slog.NewLogger(TaskTagsRepoName)
	_, err := database.Exec(createTaskTagsTable)
	assert.NoError(err, logger, "Failed to create Task Tags table")
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