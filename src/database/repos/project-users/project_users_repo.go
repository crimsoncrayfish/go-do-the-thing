package project_users_repo

import (
	"errors"
	"go-do-the-thing/src/database"
	app_errors "go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/logger"
	"go-do-the-thing/src/models"

	"github.com/jackc/pgx/v5"
)

type ProjectUsersRepo struct {
	database *database.DatabaseConnection
	logger   *logger.Logger
}

// NOTE: Depends on: [./users_repo.go, ./projects_repo.go, ../roles/roles_repo.go]
func InitRepo(database *database.DatabaseConnection) *ProjectUsersRepo {
	log := logger.NewLogger("ProjectUsersRepo")
	return &ProjectUsersRepo{
		database: database,
		logger:   log,
	}
}

func scanFromRows(rows pgx.Rows, item *models.ProjectUser) error {
	err := rows.Scan(
		&item.ProjectId,
		&item.UserId,
		&item.RoleId,
	)
	return app_errors.NewDbGenericError("failed to scan rows", err, "", nil)
}

const getAllForProject = `SELECT * FROM sp_get_all_project_users_for_project($1)`

func (r *ProjectUsersRepo) GetAllForProject(projectId int) (projectUsers []models.ProjectUser, err error) {
	rows, err := r.database.Query(getAllForProject, projectId)
	if errors.Is(err, pgx.ErrNoRows) {
		return make([]models.ProjectUser, 0), nil
	}
	if err != nil {
		r.logger.Error("failed to get all project users for projectId - query failed", "error", err, "sql", getAllForProject, "project_id", projectId)
		return nil, app_errors.NewDbReadFailed("project users", err, "for projectId", nil)
	}
	defer rows.Close()

	projectUsers = make([]models.ProjectUser, 0)
	for rows.Next() {
		var projectUser models.ProjectUser
		err = scanFromRows(rows, &projectUser)
		if err != nil {
			r.logger.Error("failed to scan row in GetAllForProject", "error", err, "project_id", projectId)
			return nil, err
		}
		projectUsers = append(projectUsers, projectUser)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error("rows.Err() in GetAllForProject", "error", err, "project_id", projectId)
		return nil, app_errors.NewDbGenericError("rows error in GetAllForProject", err, "", nil)
	}
	r.logger.Debug("GetAllForProject succeeded", "count", len(projectUsers), "project_id", projectId)
	return projectUsers, nil
}

const getAllForUser = `SELECT * FROM sp_get_all_project_users_for_user($1)`

func (r *ProjectUsersRepo) GetAllForUser(userId int) ([]models.ProjectUser, error) {
	rows, err := r.database.Query(getAllForUser, userId)
	if errors.Is(err, pgx.ErrNoRows) {
		return make([]models.ProjectUser, 0), nil
	}
	if err != nil {
		r.logger.Error("failed to get all project users for userId - query failed", "error", err, "sql", getAllForUser, "user_id", userId)
		return nil, app_errors.NewDbReadFailed("project users", err, "for userId", nil)
	}
	defer rows.Close()

	projectUsers := make([]models.ProjectUser, 0)
	for rows.Next() {
		var projectUser models.ProjectUser
		err = scanFromRows(rows, &projectUser)
		if err != nil {
			r.logger.Error("failed to scan row in GetAllForUser", "error", err, "user_id", userId)
			return nil, err
		}
		projectUsers = append(projectUsers, projectUser)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error("rows.Err() in GetAllForUser", "error", err, "user_id", userId)
		return nil, app_errors.NewDbGenericError("rows error in GetAllForUser", err, "", nil)
	}
	r.logger.Debug("GetAllForUser succeeded", "count", len(projectUsers), "user_id", userId)
	return projectUsers, nil
}

const insertProjectUser = `SELECT sp_insert_project_user($1, $2, $3)`

func (r *ProjectUsersRepo) Insert(projectId, userId, roleId int64) error {
	_, err := r.database.Exec(insertProjectUser, projectId, userId, roleId)
	if err != nil {
		r.logger.Error("failed to link user", "error", err, "sql", insertProjectUser, "project_id", projectId, "user_id", userId, "roleId", roleId)
		return app_errors.NewDbInsertFailed("project user link", err, "", nil)
	}
	r.logger.Info("User linked to project successfully", "project_id", projectId, "user_id", userId, "roleId", roleId)
	return nil
}

const updateProjectUser = `SELECT sp_update_project_user($1, $2, $3)`

func (r *ProjectUsersRepo) Update(projectId, userId, roleId int64) error {
	_, err := r.database.Exec(updateProjectUser, roleId, projectId, userId)
	if err != nil {
		r.logger.Error("failed to update link for user", "error", err, "sql", updateProjectUser, "roleId", roleId, "project_id", projectId, "user_id", userId)
		return app_errors.NewDbUpdateFailed("project user link", err, "", nil)
	}
	r.logger.Info("User project role updated successfully", "project_id", projectId, "user_id", userId, "roleId", roleId)
	return nil
}

const deleteProjectUser = `SELECT sp_delete_project_user($1, $2)`

func (r *ProjectUsersRepo) Delete(projectId, userId, roleId int64) error {
	_, err := r.database.Exec(deleteProjectUser, projectId, userId)
	if err != nil {
		r.logger.Error("failed to remove link for user", "error", err, "sql", deleteProjectUser, "project_id", projectId, "user_id", userId)
		return app_errors.NewDbDeleteFailed("project user link", err, "", nil)
	}
	r.logger.Info("User removed from project successfully", "project_id", projectId, "user_id", userId)
	return nil
}

const getAllRolesForUserProject = `SELECT * FROM sp_get_project_user_roles($1, $2)`

func (r *ProjectUsersRepo) GetProjectUserRoles(projectId, userId int64) (roleIds []int64, err error) {
	roleIds = make([]int64, 0)
	rows, err := r.database.Query(getAllRolesForUserProject, userId, projectId)
	if errors.Is(err, pgx.ErrNoRows) {
		return make([]int64, 0), nil
	}
	if err != nil {
		r.logger.Error("failed to get project_user roles", "error", err, "sql", getAllRolesForUserProject, "user_id", userId, "project_id", projectId)
		return nil, app_errors.NewDbReadFailed("project user roles", err, "", nil)
	}
	for rows.Next() {
		var val int64
		if err := rows.Scan(&val); err != nil {
			r.logger.Error("failed to get project user roles", "error", err, "sql", getAllRolesForUserProject, "user_id", userId, "project_id", projectId)
			return nil, app_errors.NewDbGenericError("failed to scan project user roles", err, "", nil)
		}
		roleIds = append(roleIds, val)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("rows.Err() in GetProjectUserRoles", "error", err, "user_id", userId, "project_id", projectId)
		return nil, app_errors.NewDbReadFailed("project user roles", err, "rows error", nil)
	}

	r.logger.Debug("GetProjectUserRoles succeeded", "count", len(roleIds), "user_id", userId, "project_id", projectId)
	return roleIds, nil
}
