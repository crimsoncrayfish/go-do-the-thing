package repos

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
)

type RolesRepo struct {
	logger   slog.Logger
	database database.DatabaseConnection
}

const RolesRepoName = "roles"

// NOTE: READONLY REPO
func initRolesRepo(database database.DatabaseConnection) *RolesRepo {
	logger := slog.NewLogger(RolesRepoName)
	_, err := database.Exec(createRolesTable)
	assert.NoError(err, logger, "Failed to create Roles table")
	_, err = database.Exec(seedRolesTable)
	assert.NoError(err, logger, "Failed to seed Roles table")
	return &RolesRepo{
		database: database,
		logger:   logger,
	}
}

const (
	createRolesTable = `CREATE TABLE IF NOT EXISTS roles (
	[id] INTEGER PRIMARY KEY,
   	[name] TEXT DEFAULT '' NOT NULL,
   	[Description] TEXT DEFAULT '' NOT NULL,
);`
	seedRolesTable = `SOME SQL HERE TO SEED THE ROLES`
	getAllRoles    = `SOME SQL HERE TO GET ALL ROLES`
)
