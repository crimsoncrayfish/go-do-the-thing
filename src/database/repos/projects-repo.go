package repos

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
)

type ProjectsRepo struct {
	database database.DatabaseConnection
}

// NOTE: Depends on: [./users-repo.go]
func initProjectsRepo(database database.DatabaseConnection) *ProjectsRepo {
	logger := slog.NewLogger("projects repo")
	_, err := database.Exec(createProjectsTable)
	assert.NoError(err, logger, "Failed to create Projects table")

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

	getProjects   = `SELECT [],[],[] FROM projects WHERE owner = ?`
	getProject    = `SELECT [],[],[] FROM projects WHERE owner = ? AND id = ?`
	insertProject = `INSERT INTO projects `
	updateProject = `UPDATE projects SET VALUES() WHERE id = ?`
	deleteProject = `UPDATE projects SET [is_deleted] = 1, [modified_by] = ?, [modified_date] = ? WHERE id = ?`
)
