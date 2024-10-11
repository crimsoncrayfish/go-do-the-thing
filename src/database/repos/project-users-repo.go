package repos

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
)

type ProjectUsersRepo struct {
	logger   slog.Logger
	database database.DatabaseConnection
}

const ProjectUsersRepoName = "tags"

// NOTE: READONLY REPO
func InitProjectUsersRepo(database database.DatabaseConnection) (*ProjectUsersRepo, error) {
	logger := slog.NewLogger(ProjectUsersRepoName)
	_, err := database.Exec(createProjectUsersTable)
	assert.NoError(err, logger, "Failed to create ProjectUsers table")
	_, err = database.Exec(seedProjectUsersTable)
	assert.NoError(err, logger, "Failed to seed ProjectUsers table")
	return &ProjectUsersRepo{
		database: database,
		logger:   logger,
	}, nil
}

const (
	createProjectUsersTable = `CREATE TABLE IF NOT EXISTS tags (
	[id] INTEGER PRIMARY KEY,
   	[name] TEXT DEFAULT '' NOT NULL,
);`
	seedProjectUsersTable = `SOME SQL HERE TO SEED THE TAGS`
	getAllProjectUsers    = `SOME SQL HERE TO GET ALL TAGS`
)
