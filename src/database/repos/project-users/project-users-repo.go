package project_users_repo

import (
	"database/sql"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
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

func scanFromRow(row *sql.Row, item *models.ProjectUser) error {
	return row.Scan(
		&item.ProjectId,
		&item.UserId,
		&item.RoleId,
	)
}

func scanFromRows(rows *sql.Rows, item *models.ProjectUser) error {
	return rows.Scan(
		&item.ProjectId,
		&item.UserId,
		&item.RoleId,
	)
}

const getAllForProject = `SELECT [project_id], [user_id], [role_id] FROM project_users WHERE project_id = ?`

func (r *ProjectUsersRepo) GetAllForProject(projectId int) ([]models.ProjectUser, error) {
	rows, err := r.database.Query(getAllForProject, projectId)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
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
		return nil, err
	}
	return projectUsers, nil
}

const getAllForUser = `SELECT [project_id], [user_id], [role_id] FROM project_users WHERE user_id = ?`

func (r *ProjectUsersRepo) GetAllForUser(userId int) ([]models.ProjectUser, error) {
	rows, err := r.database.Query(getAllForUser, userId)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
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
		return nil, err
	}
	return projectUsers, nil
}

const insertProjectUser = `INSERT INTO project_users (project_id, user_id, role_id) VALUES (?, ?, ?)`

func (r *ProjectUsersRepo) Insert(projectId, userId, roleId int64) error {
	_, err := r.database.Exec(insertProjectUser, projectId, userId, roleId)

	return err
}

const updateProjectUser = `UPDATE project_users SET [role_id] = ? WHERE [project_id] = ? AND [user_id] = ?`

func (r *ProjectUsersRepo) Update(projectId, userId, roleId int64) error {
	_, err := r.database.Exec(updateProjectUser, roleId, projectId, userId)

	return err
}

const deleteProjectUser = `DELETE FROM project_users WHERE [project_id] = ? AND [user_id] = ?`

func (r *ProjectUsersRepo) Delete(projectId, userId, roleId int64) error {
	_, err := r.database.Exec(deleteProjectUser, projectId, userId)

	return err
}
