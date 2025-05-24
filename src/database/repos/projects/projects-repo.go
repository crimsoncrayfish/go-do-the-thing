package projects_repo

import (
	"database/sql"
	"fmt"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"
)

type ProjectsRepo struct {
	database database.DatabaseConnection
	logger   slog.Logger
}

var repoName = "ProjectsRepo"

// NOTE: Depends on: [../project-users/project-users-repo.go, ../users/users-repo.go]
func InitRepo(database database.DatabaseConnection) *ProjectsRepo {
	_, err := database.Exec(createProjectsTable)
	assert.NoError(err, repoName, "Failed to create Projects table")

	return &ProjectsRepo{
		database: database,
		logger:   slog.NewLogger(repoName),
	}
}

const createProjectsTable = `CREATE TABLE IF NOT EXISTS projects (
	[id] INTEGER PRIMARY KEY,
   	[name] TEXT DEFAULT '' NOT NULL,
   	[description] TEXT,
	[owner] INT,
	[start_date] INT,
	[due_date] INT,
	[created_by] TEXT,
	[created_date] INT,
	[modified_by] TEXT,
	[modified_date] INT,
	[is_complete] INT,
	[is_deleted] INT,
	FOREIGN KEY (owner) REFERENCES users(id),
	FOREIGN KEY (created_by) REFERENCES users(id),
	FOREIGN KEY (modified_by) REFERENCES users(id)
);`

func scanFromRow(row *sql.Row, item *models.Project) error {
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

func scanFromRows(rows *sql.Rows, item *models.Project) error {
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

const getProjectsByUser = `
	SELECT 
		projects.[id],projects.[name],projects.[description],projects.[owner],projects.[start_date],projects.[due_date],projects.[created_by],projects.[created_date],projects.[modified_by],projects.[modified_date],projects.[is_complete],projects.[is_deleted] 
	FROM projects 
	JOIN project_users
		ON projects.id = project_users.project_id
	WHERE project_users.user_id = ?
	AND projects.is_deleted = 0`

func (r *ProjectsRepo) GetProjects(user_id int64) ([]models.Project, error) {
	rows, err := r.database.Query(getProjectsByUser, user_id)
	if err != nil {
		return nil, fmt.Errorf("failed to get project list, query failed: %w", err)
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			err = fmt.Errorf("failed to close rows object: %w", err)
		}
	}(rows)

	items := make([]models.Project, 0)
	for rows.Next() {
		item := models.Project{}

		err = scanFromRows(rows, &item)
		if err != nil {
			return nil, fmt.Errorf("failed to get project list, scan failed: %w", err)
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("failed to get project list, row error: %w", err)
	}
	return items, nil
}

const getProject = `
	SELECT 
		[id],[name],[description],[owner],[start_date],[due_date],[created_by],[created_date],[modified_by],[modified_date],[is_complete],[is_deleted]
	FROM projects
	JOIN project_users
		ON projects.id = project_users.project_id
	WHERE 
		projects.id = ?`

func (r *ProjectsRepo) GetProject(projectId int64) (*models.Project, error) {
	row := r.database.QueryRow(getProject, projectId)
	temp := &models.Project{}
	err := scanFromRow(row, temp)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return temp, nil
}

const deleteProject = `
	UPDATE projects 
	SET [is_deleted] = 1, [modified_by] = ?, [modified_date] = ?
	WHERE id = ?`

func (r *ProjectsRepo) DeleteProject(id, currentUser int64) error {
	_, err := r.database.Exec(deleteProject, currentUser, database.SqLiteNow(), id)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}

const getProjectCount = `SELECT COUNT(id) FROM items WHERE [is_deleted]=0 AND [assigned_to]=?`

func (r *ProjectsRepo) GetProjectCount(currentUser int64) (count int64, err error) {
	row := r.database.QueryRow(getProjectCount, currentUser)
	var temp int64
	err = row.Scan(&temp)
	if err != nil {
		return 0, fmt.Errorf("failed to get project count: %w", err)
	}
	return temp, nil
}

const updateProject = `
	UPDATE projects
	SET	
		[name] = ?,
		[description] = ?,
		[owner] = ?,
		[start_date] = ?,
		[due_date] = ?,
		[modified_by] = ?,
		[modified_date] = ?
	WHERE id = ?`

func (r *ProjectsRepo) UpdateProject(project models.Project) (err error) {
	_, err = r.database.Exec(updateProject,
		project.Name,
		project.Description,
		project.Owner,
		project.StartDate,
		project.DueDate,
		project.ModifiedBy,
		project.ModifiedDate,
		project.Id,
	)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}

const insertProject = `
	INSERT INTO projects 
	(
		[name],
		[description],
		[owner],
		[start_date],
		[due_date],
		[created_by],
		[created_date],
		[modified_by],
		[modified_date],
		[is_complete],
		[is_deleted]
	)
	VALUES (?,?,?,?,?,?,?,?,?,?,?)`

func (r *ProjectsRepo) Insert(currentUser int64, project models.Project) (id int64, err error) {
	project.AssertHealthyNew()
	assert.NotEqual(currentUser, 0, repoName, "currentUser")

	result, err := r.database.Exec(
		insertProject,
		project.Name,
		project.Description,
		project.Owner,
		project.StartDate,
		project.DueDate,
		currentUser,
		database.SqLiteNow(),
		currentUser,
		database.SqLiteNow(),
		project.IsComplete,
		false,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create new project: %w", err)
	}

	id, err = result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to create new project: %w", err)
	}

	return id, nil
}
