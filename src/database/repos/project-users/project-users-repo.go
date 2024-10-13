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

// NOTE: Depends on: [./users-repo.go, ./projects-repo.go, ../roles/roles-repo.go]
func InitRepo(database database.DatabaseConnection) *ProjectUsersRepo {
	logger := slog.NewLogger("project users repo")
	_, err := database.Exec(createProjectUsersTable)
	assert.NoError(err, logger, "Failed to create ProjectUsers table")

	return &ProjectUsersRepo{
		database: database,
		logger:   logger,
	}
}

const (
	createProjectUsersTable = `CREATE TABLE IF NOT EXISTS project_users (
	[project_id] INTEGER,
	[user_id] INTEGER,
	[role_id] INTEGER,
	FOREIGN KEY (project_id) REFERENCES projects(id)
	FOREIGN KEY (user_id) REFERENCES users(id),
	FOREIGN KEY (role_id) REFERENCES roles(id)
);`
	getAllForProject = `SELECT [project_id], [user_id], [role_id] FROM project_users WHERE project_id = ?`
	getAllForUser    = `SELECT [project_id], [user_id], [role_id] FROM project_users WHERE user_id = ?`
	insert           = `INSERT INTO project_users (project_id, user_id, role_id) VALUES (?, ?, ?)`
	update           = `UPDATE project_users SET [role_id] = ? WHERE [project_id] = ? AND [user_id] = ?`
	delete           = `DELETE FROM project_users WHERE [project_id] = ? AND [user_id] = ?`
	//assertUserIsInProject = ``
	//assertUserHasRole     = ``
)
