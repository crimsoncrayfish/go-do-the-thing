package projects_repo

import (
	"fmt"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"
	"time"

	"github.com/jackc/pgx/v5"
)

type ProjectsRepo struct {
	database database.DatabaseConnection
	logger   slog.Logger
}

var repoName = "ProjectsRepo"

// NOTE: Depends on: [../project-users/project_users_repo.go, ../users/users_repo.go]
func InitRepo(database database.DatabaseConnection) *ProjectsRepo {
	//TODO: Cleanup
	//_, err := database.Exec(createProjectsTable)
	//assert.NoError(err, repoName, "Failed to create Projects table")

	logger := slog.NewLogger(repoName)
	return &ProjectsRepo{
		database: database,
		logger:   logger,
	}
}

func scanFromRow(row pgx.Row, item *models.Project) error {
	return row.Scan(
		&item.Id,
		&item.Name,
		&item.Description,
		&item.Owner,
		&item.StartDate,
		&item.DueDate,
		&item.CreatedBy,
		&item.CreatedDate,
		&item.ModifiedBy,
		&item.ModifiedDate,
		&item.IsComplete,
		&item.IsDeleted,
	)
}

func scanFromRows(rows pgx.Rows, item *models.Project) error {
	return rows.Scan(
		&item.Id,
		&item.Name,
		&item.Description,
		&item.Owner,
		&item.StartDate,
		&item.DueDate,
		&item.CreatedBy,
		&item.CreatedDate,
		&item.ModifiedBy,
		&item.ModifiedDate,
		&item.IsComplete,
		&item.IsDeleted,
	)
}

const getProjectsByUser = `SELECT * FROM sp_get_projects_by_user($1)`

func (r *ProjectsRepo) GetProjects(user_id int64) ([]models.Project, error) {
	r.logger.Debug("GetProjects called - sql: %s, params: %v", getProjectsByUser, user_id)
	rows, err := r.database.Query(getProjectsByUser, user_id)
	if err != nil {
		r.logger.Error(err, "failed to get project list, query failed - sql: %s, params: %v", getProjectsByUser, user_id)
		return nil, errors.New(errors.ErrDBReadFailed, "failed to get project list, query failed: %w", err)
	}
	defer rows.Close()

	items := make([]models.Project, 0)
	for rows.Next() {
		var item models.Project
		err = scanFromRows(rows, &item)
		if err != nil {
			r.logger.Error(err, "failed to get project list, scan failed - params: %v", user_id)
			return nil, errors.New(errors.ErrDBGenericError, "failed to get project list, scan failed: %w", err)
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error(err, "failed to get project list, row error - params: %v", user_id)
		return nil, errors.New(errors.ErrDBGenericError, "failed to get project list, row error: %w", err)
	}
	r.logger.Debug("GetProjects succeeded - count: %d, params: %v", len(items), user_id)
	return items, nil
}

const getProject = `SELECT * FROM sp_get_project($1)`

func (r *ProjectsRepo) GetProject(projectId int64) (*models.Project, error) {
	r.logger.Debug("GetProject called - sql: %s, params: %v", getProject, projectId)
	row := r.database.QueryRow(getProject, projectId)
	temp := &models.Project{}
	err := scanFromRow(row, temp)
	if err != nil {
		r.logger.Error(err, "failed to get project - sql: %s, params: %v", getProject, projectId)
		return nil, errors.New(errors.ErrDBReadFailed, "failed to get project: %w", err)
	}
	r.logger.Debug("GetProject succeeded - id: %v", projectId)
	return temp, nil
}

const deleteProject = `SELECT sp_delete_project($1, $2, $3)`

func (r *ProjectsRepo) DeleteProject(id, currentUser int64) error {
	r.logger.Debug("DeleteProject called - sql: %s, params: %v", deleteProject, []interface{}{currentUser, time.Now(), id})
	_, err := r.database.Exec(deleteProject, currentUser, time.Now(), id)
	if err != nil {
		r.logger.Error(err, "failed to update project - sql: %s, params: %v", deleteProject, []interface{}{currentUser, time.Now(), id})
		return errors.New(errors.ErrDBUpdateFailed, "failed to update project: %w", err)
	}
	r.logger.Info("Project deleted successfully - id: %d", id)
	return nil
}

const getProjectCount = `SELECT sp_get_project_count($1)`

func (r *ProjectsRepo) GetProjectCount(currentUser int64) (count int64, err error) {
	row := r.database.QueryRow(getProjectCount, currentUser)
	var temp int64
	err = row.Scan(&temp)
	if err != nil {
		return 0, fmt.Errorf("failed to get project count: %w", err)
	}
	return temp, nil
}

const updateProject = `SELECT sp_update_project($1, $2, $3, $4, $5, $6, $7, $8)`

func (r *ProjectsRepo) UpdateProject(project models.Project) (err error) {
	r.logger.Debug("UpdateProject called - sql: %s, params: %+v", updateProject, project)
	_, err = r.database.Exec(updateProject,
		project.Id,
		project.Name,
		project.Description,
		project.Owner,
		project.StartDate,
		project.DueDate,
		project.ModifiedBy,
		time.Now(),
	)
	if err != nil {
		r.logger.Error(err, "failed to update project - sql: %s, params: %+v", updateProject, project)
		return errors.New(errors.ErrDBUpdateFailed, "failed to update project: %w", err)
	}
	r.logger.Info("Project updated successfully - id: %d, name: %s", project.Id, project.Name)
	return nil
}

const insertProject = `SELECT sp_insert_project($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

func (r *ProjectsRepo) Insert(project models.Project) (id int64, err error) {
	r.logger.Debug("InsertProject called - sql: %s, params: %+v", insertProject, project)
	project.AssertHealthyNew()

	err = r.database.QueryRow(
		insertProject,
		project.Name,
		project.Description,
		project.Owner,
		project.StartDate,
		project.DueDate,
		project.CreatedBy,
		time.Now(),
		project.ModifiedBy,
		time.Now(),
		project.IsComplete,
		project.IsDeleted,
	).Scan(&id)
	if err != nil {
		r.logger.Error(err, "failed to create new project - sql: %s, params: %+v", insertProject, project)
		return 0, errors.New(errors.ErrDBReadFailed, "failed to create new project: %w", err)
	}
	r.logger.Info("Project created successfully - id: %d, name: %s, owner: %d", id, project.Name, project.Owner)
	return id, nil
}
