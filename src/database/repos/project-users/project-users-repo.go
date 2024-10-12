package project_users_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
)

type ProjectUsersRepo struct {
	logger   slog.Logger
	database database.DatabaseConnection
}

// NOTE: Depends on: [./users-repo.go, ./projects-repo.go]
func InitProjectUsersRepo(database database.DatabaseConnection) *ProjectUsersRepo {
	logger := slog.NewLogger("project users repo")
	_, err := database.Exec(createProjectUsersTable)
	assert.NoError(err, logger, "Failed to create ProjectUsers table")

	return &ProjectUsersRepo{
		database: database,
		logger:   logger,
	}
}

const (
	createProjectUsersTable = `CREATE TABLE IF NOT EXISTS tags (
	[id] INTEGER PRIMARY KEY,
   	[name] TEXT DEFAULT '' NOT NULL,
);`
	getAllProjectUsers = `SOME SQL HERE TO GET ALL TAGS`
)
