package project_users_repo

import (
	"errors"
	"go-do-the-thing/src/database"
	app_errors "go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"

	"github.com/jackc/pgx/v5"
)

type ProjectUsersRepo struct {
	database database.DatabaseConnection
	logger   slog.Logger
}

var repoName = "Project Users Repo"

// NOTE: Depends on: [./users_repo.go, ./projects_repo.go, ../roles/roles_repo.go]
func InitRepo(database database.DatabaseConnection) *ProjectUsersRepo {
	logger := slog.NewLogger(repoName)
	//TODO: Cleanup
	//_, err := database.Exec(createProjectUsersTable)
	//assert.NoError(err, repoName, "Failed to create ProjectUsers table")

	return &ProjectUsersRepo{
		database: database,
		logger:   logger,
	}
}

func scanFromRows(rows pgx.Rows, item *models.ProjectUser) error {
	err := rows.Scan(
		&item.ProjectId,
		&item.UserId,
		&item.RoleId,
	)
	return app_errors.New(app_errors.ErrDBGenericError, "failed to scan rows: %w", err)
}

const getAllForProject = `SELECT * FROM sp_get_all_project_users_for_project($1)`

func (r *ProjectUsersRepo) GetAllForProject(projectId int) (projectUsers []models.ProjectUser, err error) {
	r.logger.Debug("GetAllForProject called - sql: %s, params: %d", getAllForProject, projectId)
	rows, err := r.database.Query(getAllForProject, projectId)
	if errors.Is(err, pgx.ErrNoRows) {
		return make([]models.ProjectUser, 0), nil
	}
	if err != nil {
		r.logger.Error(err, "failed to get all project users for projectId - query failed - sql: %s, params: %d", getAllForProject, projectId)
		return nil, app_errors.New(app_errors.ErrDBReadFailed, "failed to get all project users for projectId - query failed: %w", err)
	}
	defer rows.Close()

	projectUsers = make([]models.ProjectUser, 0)
	for rows.Next() {
		var projectUser models.ProjectUser
		err = scanFromRows(rows, &projectUser)
		if err != nil {
			r.logger.Error(err, "failed to scan row in GetAllForProject - params: %d", projectId)
			return nil, err
		}
		projectUsers = append(projectUsers, projectUser)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error(err, "rows.Err() in GetAllForProject - params: %d", projectId)
		return nil, app_errors.New(app_errors.ErrDBGenericError, "some error contained in the rows: %w", err)
	}
	r.logger.Debug("GetAllForProject succeeded - count: %d, params: %d", len(projectUsers), projectId)
	return projectUsers, nil
}

const getAllForUser = `SELECT * FROM sp_get_all_project_users_for_user($1)`

func (r *ProjectUsersRepo) GetAllForUser(userId int) ([]models.ProjectUser, error) {
	r.logger.Debug("GetAllForUser called - sql: %s, params: %d", getAllForUser, userId)
	rows, err := r.database.Query(getAllForUser, userId)
	if errors.Is(err, pgx.ErrNoRows) {
		return make([]models.ProjectUser, 0), nil
	}
	if err != nil {
		r.logger.Error(err, "failed to get all project users for userId - query failed - sql: %s, params: %d", getAllForUser, userId)
		return nil, app_errors.New(app_errors.ErrDBReadFailed, "failed to get all project users for userId - query failed: %w", err)
	}
	defer rows.Close()

	projectUsers := make([]models.ProjectUser, 0)
	for rows.Next() {
		var projectUser models.ProjectUser
		err = scanFromRows(rows, &projectUser)
		if err != nil {
			r.logger.Error(err, "failed to scan row in GetAllForUser - params: %d", userId)
			return nil, err
		}
		projectUsers = append(projectUsers, projectUser)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error(err, "rows.Err() in GetAllForUser - params: %d", userId)
		return nil, app_errors.New(app_errors.ErrDBGenericError, "some error contained in the rows: %w", err)
	}
	r.logger.Debug("GetAllForUser succeeded - count: %d, params: %d", len(projectUsers), userId)
	return projectUsers, nil
}

const insertProjectUser = `SELECT sp_insert_project_user($1, $2, $3)`

func (r *ProjectUsersRepo) Insert(projectId, userId, roleId int64) error {
	r.logger.Debug("Insert called - sql: %s, params: %v", insertProjectUser, []interface{}{projectId, userId, roleId})
	_, err := r.database.Exec(insertProjectUser, projectId, userId, roleId)
	if err != nil {
		r.logger.Error(err, "failed to link user - sql: %s, params: %v", insertProjectUser, []interface{}{projectId, userId, roleId})
		return app_errors.New(app_errors.ErrDBInsertFailed, "failed to link user (%d) to project (%d): %w", userId, projectId, err)
	}
	r.logger.Info("User linked to project successfully - projectId: %d, userId: %d, roleId: %d", projectId, userId, roleId)
	return nil
}

const updateProjectUser = `SELECT sp_update_project_user($1, $2, $3)`

func (r *ProjectUsersRepo) Update(projectId, userId, roleId int64) error {
	r.logger.Debug("Update called - sql: %s, params: %v", updateProjectUser, []interface{}{roleId, projectId, userId})
	_, err := r.database.Exec(updateProjectUser, roleId, projectId, userId)
	if err != nil {
		r.logger.Error(err, "failed to update link for user - sql: %s, params: %v", updateProjectUser, []interface{}{roleId, projectId, userId})
		return app_errors.New(app_errors.ErrDBUpdateFailed, "failed to update link for user (%d) and project (%d): %w", userId, projectId, err)
	}
	r.logger.Info("User project role updated successfully - projectId: %d, userId: %d, roleId: %d", projectId, userId, roleId)
	return nil
}

const deleteProjectUser = `SELECT sp_delete_project_user($1, $2)`

func (r *ProjectUsersRepo) Delete(projectId, userId, roleId int64) error {
	r.logger.Debug("Delete called - sql: %s, params: %v", deleteProjectUser, []interface{}{projectId, userId})
	_, err := r.database.Exec(deleteProjectUser, projectId, userId)
	if err != nil {
		r.logger.Error(err, "failed to remove link for user - sql: %s, params: %v", deleteProjectUser, []interface{}{projectId, userId})
		return app_errors.New(app_errors.ErrDBDeleteFailed, "failed to remove link for user (%d) and project (%d): %w", userId, projectId, err)
	}
	r.logger.Info("User removed from project successfully - projectId: %d, userId: %d", projectId, userId)
	return nil
}

const getAllRolesForUserProject = `SELECT * FROM sp_get_project_user_roles($1, $2)`

func (r *ProjectUsersRepo) GetProjectUserRoles(projectId, userId int64) (roleIds []int64, err error) {
	r.logger.Debug("GetProjectUserRoles called - sql: %s, params: %v", getAllRolesForUserProject, []interface{}{userId, projectId})
	roleIds = make([]int64, 0)
	rows, err := r.database.Query(getAllRolesForUserProject, userId, projectId)
	if errors.Is(err, pgx.ErrNoRows) {
		return make([]int64, 0), nil
	}
	if err != nil {
		r.logger.Error(err, "failed to get project_user roles - sql: %s, params: %v", getAllRolesForUserProject, []interface{}{userId, projectId})
		return nil, app_errors.New(app_errors.ErrDBReadFailed, "failed to get project_user roles: %w", err)
	}
	for rows.Next() {
		var val int64
		if err := rows.Scan(&val); err != nil {
			r.logger.Error(err, "failed to get project user roles - sql: %s, params: %v", getAllRolesForUserProject, []interface{}{userId, projectId})
			return nil, app_errors.New(app_errors.ErrDBGenericError, "failed to get project user roles: %w", err)
		}
		roleIds = append(roleIds, val)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error(err, "rows.Err() in GetProjectUserRoles - params: %v", []interface{}{userId, projectId})
		return nil, app_errors.New(app_errors.ErrDBReadFailed, "failed to get project_user roles: %w", err)
	}

	r.logger.Debug("GetProjectUserRoles succeeded - count: %d, params: %v", len(roleIds), []interface{}{userId, projectId})
	return roleIds, nil
}
