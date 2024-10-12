package repos

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
)

type ProjectTagsRepo struct {
	database database.DatabaseConnection
}

const ProjectTagsRepoName = "project-tags"

func initProjectTagsRepo(database database.DatabaseConnection) *ProjectTagsRepo {
	logger := slog.NewLogger(ProjectTagsRepoName)
	_, err := database.Exec(createProjectTagsTable)
	assert.NoError(err, logger, "Failed to create Project Tags table")
	return &ProjectTagsRepo{
		database: database,
	}
}

const (
	createProjectTagsTable = `CREATE TABLE IF NOT EXISTS project_tags (
	[project_id] INTEGER,
	[tag_id] INTEGER,
);`
)
