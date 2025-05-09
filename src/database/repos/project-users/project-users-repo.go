package project_users_repo

import (
	"database/sql"
	"errors"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	app_errors "go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/models"
)

type ProjectUsersRepo struct {
	database database.DatabaseConnection
}

var repoName = "Project Users Repo"

// NOTE: Depends on: [./users-repo.go, ./projects-repo.go, ../roles/roles-repo.go]
func InitRepo(database database.DatabaseConnection) *ProjectUsersRepo {
	_, err := database.Exec(createProjectUsersTable)
	assert.NoError(err, repoName, "Failed to create ProjectUsers table")

	return &ProjectUsersRepo{
		database: database,
	}
}

const createProjectUsersTable = `CREATE TABLE IF NOT EXISTS project_users (
	[project_id] INTEGER,
	[user_id] INTEGER,
	[role_id] INTEGER,
	FOREIGN KEY (project_id) REFERENCES projects(id)
	FOREIGN KEY (user_id) REFERENCES users(id),
	FOREIGN KEY (role_id) REFERENCES roles(id)
);`

func scanFromRows(rows *sql.Rows, item *models.ProjectUser) error {
	err := rows.Scan(
		&item.ProjectId,
		&item.UserId,
		&item.RoleId,
	)
	return app_errors.New(app_errors.ErrDBGenericError, "failed to scan rows: %w", err)
}

const getAllForProject = `SELECT [project_id], [user_id], [role_id] FROM project_users WHERE project_id = ?`

func (r *ProjectUsersRepo) GetAllForProject(projectId int) (projectUsers []models.ProjectUser, err error) {
	rows, err := r.database.Query(getAllForProject, projectId)
	if errors.Is(err, sql.ErrNoRows) {
		return make([]models.ProjectUser, 0), nil
	}
	if err != nil {
		return nil, app_errors.New(app_errors.ErrDBReadFailed, "failed to get all project users for projectId - query failed: %w", err)
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		err = app_errors.New(app_errors.ErrDBGenericError, "failed to get all project users for userId - row error: %w", err)
	}(rows)

	projectUsers = make([]models.ProjectUser, 0)
	for rows.Next() {
		projectUser := models.ProjectUser{}

		err = scanFromRows(rows, &projectUser)
		if err != nil {
			return nil, err
		}
		projectUsers = append(projectUsers, projectUser)
	}
	err = rows.Err()
	if err != nil {
		return nil, app_errors.New(app_errors.ErrDBGenericError, "failed to get all project users for projectId - row error: %w", err)
	}
	return projectUsers, nil
}

const getAllForUser = `SELECT [project_id], [user_id], [role_id] FROM project_users WHERE user_id = ?`

func (r *ProjectUsersRepo) GetAllForUser(userId int) ([]models.ProjectUser, error) {
	rows, err := r.database.Query(getAllForUser, userId)
	if errors.Is(err, sql.ErrNoRows) {
		return make([]models.ProjectUser, 0), nil
	}
	if err != nil {
		return nil, app_errors.New(app_errors.ErrDBReadFailed, "failed to get all project users for userId - query failed: %w", err)
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		err = app_errors.New(app_errors.ErrDBGenericError, "failed to get all project users for userId - row error: %w", err)
	}(rows)

	projectUsers := make([]models.ProjectUser, 0)
	for rows.Next() {
		projectUser := models.ProjectUser{}

		err = scanFromRows(rows, &projectUser)
		if err != nil {
			return nil, err
		}
		projectUsers = append(projectUsers, projectUser)
	}
	err = rows.Err()
	if err != nil {
		return nil, app_errors.New(app_errors.ErrDBGenericError, "failed to get all project users for userId - row error: %w", err)
	}
	return projectUsers, nil
}

const insertProjectUser = `INSERT INTO project_users (project_id, user_id, role_id) VALUES (?, ?, ?)`

func (r *ProjectUsersRepo) Insert(projectId, userId, roleId int64) (int64, error) {
	result, err := r.database.Exec(insertProjectUser, projectId, userId, roleId)
	if err != nil {
		return 0, app_errors.New(app_errors.ErrDBInsertFailed, "failed to link user (%d) to project (%d): %w", userId, projectId, err)
	}
	inserted_id, err := result.LastInsertId()
	if err != nil {
		return 0, app_errors.New(app_errors.ErrDBGenericError, "failed to link user (%d) to project (%d): %w", userId, projectId, err)
	}

	return inserted_id, nil
}

const updateProjectUser = `UPDATE project_users SET [role_id] = ? WHERE [project_id] = ? AND [user_id] = ?`

func (r *ProjectUsersRepo) Update(projectId, userId, roleId int64) error {
	_, err := r.database.Exec(updateProjectUser, roleId, projectId, userId)
	if err != nil {
		return app_errors.New(app_errors.ErrDBUpdateFailed, "failed to update link for user (%d) and project (%d): %w", userId, projectId, err)
	}

	return nil
}

const deleteProjectUser = `DELETE FROM project_users WHERE [project_id] = ? AND [user_id] = ?`

func (r *ProjectUsersRepo) Delete(projectId, userId, roleId int64) error {
	_, err := r.database.Exec(deleteProjectUser, projectId, userId)
	if err != nil {
		return app_errors.New(app_errors.ErrDBDeleteFailed, "failed to remove link for user (%d) and project (%d): %w", userId, projectId, err)
	}

	return nil
}

const (
	getAllRolesForUserProject = `
	SELECT 
		[role_id]
	FROM project_users
	WHERE user_id = ? AND
	[project_id] = ?`
)

func (r *ProjectUsersRepo) GetProjectUserRoles(projectId, userId int64) (roleIds []int64, err error) {
	roleIds = make([]int64, 0)
	rows, err := r.database.Query(getAllRolesForUserProject, userId, projectId)
	if errors.Is(err, sql.ErrNoRows) {
		return roleIds, nil
	}
	if err != nil {
		return nil, app_errors.New(app_errors.ErrDBReadFailed, "failed to get project_user roles: %w", err)
	}
	for rows.Next() {
		var val int64
		if err := rows.Scan(&val); err != nil {
			return nil, app_errors.New(app_errors.ErrDBGenericError, "failed to get project user roles: %w", err)
		}
		roleIds = append(roleIds, val)
	}

	if err := rows.Err(); err != nil {
		return nil, app_errors.New(app_errors.ErrDBReadFailed, "failed to get project_user roles: %w", err)
	}

	return roleIds, nil
}
