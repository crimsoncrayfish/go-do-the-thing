package repos

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
)

type ProjectsRepo struct {
	database database.DatabaseConnection
}

const ProjectsRepoName = "projects"

func InitProjectsRepo(database database.DatabaseConnection) (*ProjectsRepo, error) {
	logger := slog.NewLogger(ProjectUsersRepoName)
	_, err := database.Exec(createProjectsTable)
	assert.NoError(err, logger, "Failed to create Projects table")
	return &ProjectsRepo{
		database: database,
	}, nil
}

const (
	createProjectsTable = `CREATE TABLE IF NOT EXISTS projects (
	[id] INTEGER PRIMARY KEY,
   	[name] TEXT DEFAULT '' NOT NULL,
   	[description] TEXT,
);`
	createProjectTagsTable = `CREATE TABLE IF NOT EXISTS project_tags (
	[project_id] INTEGER,
	[tag_id] INTEGER,
);`
)
