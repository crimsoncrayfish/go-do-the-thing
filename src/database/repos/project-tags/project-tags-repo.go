package project_tags_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
)

type ProjectTagsRepo struct {
	database database.DatabaseConnection
}

// NOTE: Depends on: [./tags-repo.go, ./projects-repo.go]
func InitRepo(database database.DatabaseConnection) *ProjectTagsRepo {
	logger := slog.NewLogger("project tags repo")
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
	insertProjectTag = `INSERT INTO project_tags (project_id, tag_id) VALUES (?, ?)`
	deleteTag        = `DELETE FROM project_tags WHERE [tag_id] = ?`
	deleteProject    = `DELETE FROM project_tags WHERE [project_id] = ?`
	deleteProjectTag = `DELETE FROM project_tags WHERE [tag_id] = ? AND [project_id] = ?`
)
