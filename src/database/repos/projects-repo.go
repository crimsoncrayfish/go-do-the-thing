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

func initProjectsRepo(database database.DatabaseConnection) *ProjectsRepo {
	logger := slog.NewLogger(ProjectUsersRepoName)
	_, err := database.Exec(createProjectsTable)
	assert.NoError(err, logger, "Failed to create Projects table")
	_, err = database.Exec(createProjectTagsTable)
	assert.NoError(err, logger, "Failed to create Project Tags table")
	return &ProjectsRepo{
		database: database,
	}
}

const (
	createProjectsTable = `CREATE TABLE IF NOT EXISTS projects (
	[id] INTEGER PRIMARY KEY,
   	[name] TEXT DEFAULT '' NOT NULL,
   	[description] TEXT,
);`

	insertProject = `INSERT INTO projects `
)
