package projects_repo

import (
	"database/sql"
	"fmt"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"
	"strings"
)

type ProjectsRepo struct {
	database database.DatabaseConnection
}

// NOTE: Depends on: [./users-repo.go]
func InitRepo(database database.DatabaseConnection) *ProjectsRepo {
	logger := slog.NewLogger("projects repo")
	_, err := database.Exec(createProjectsTable)
	assert.NoError(err, logger, "Failed to create Projects table")

	return &ProjectsRepo{
		database: database,
	}
}

const (
	createProjectsTable = `CREATE TABLE IF NOT EXISTS projects (
	[id] INTEGER PRIMARY KEY,
   	[name] TEXT DEFAULT '' NOT NULL,
   	[description] TEXT,
	[owner] INT,
	[start_date] TEXT,
	[due_date] TEXT,
	[created_by] TEXT,
	[created_date] TEXT,
	[modified_by] TEXT,
	[modified_date] TEXT,
	[is_complete] INT,
	[is_deleted] INT,
	FOREIGN KEY (owner) REFERENCES users(id),
	FOREIGN KEY (created_by) REFERENCES users(id),
	FOREIGN KEY (modified_by) REFERENCES users(id),
);`

	getProjectsByOwner = `
	SELECT 
		[id],[name],[description],[owner],[start_date],[due_date],[created_by],[create_date],[modified_by],[modified_date],[is_complete],[is_deleted] 
	FROM projects 
	WHERE owner = ?`
	getProjects = `
	SELECT 
		[id],[name],[description],[owner],[start_date],[due_date],[created_by],[create_date],[modified_by],[modified_date],[is_complete],[is_deleted] 
	FROM projects 
	WHERE id IN (?)`
	getProject = `
	SELECT 
		[id],[name],[description],[owner],[start_date],[due_date],[created_by],[create_date],[modified_by],[modified_date],[is_complete],[is_deleted]
	FROM projects
	WHERE id = ?`
	insertProject = `
	INSERT INTO projects 
	([id],[name],[description],[owner],[start_date],[due_date],[created_by],[create_date],[modified_by],[modified_date],[is_complete],[is_deleted])
	VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`
	updateProject = `
	UPDATE projects
	SET	
		[id] = ?,
		[name] = ?,
		[description] = ?,
		[owner] = ?,
		[start_date] = ?,
		[due_date] = ?,
		[modified_by] = ?,
		[modified_date] = ?,
		[is_complete] = ?,
		[is_deleted] = ?
	WHERE id = ?`
	deleteProject = `
	UPDATE projects 
	SET [is_deleted] = 1, [modified_by] = ?, [modified_date] = ?
	WHERE id = ?`
)

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

func (r *ProjectsRepo) GetProjects(idlist []int64) ([]models.Project, error) {
	ids := strings.Trim(strings.Join(strings.Split(fmt.Sprint(idlist), " "), ","), "[]")
	rows, err := r.database.Query(getProjects, ids)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
	}(rows)

	items := make([]models.Project, 0)
	for rows.Next() {
		item := models.Project{}

		err = scanFromRows(rows, &item)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *ProjectsRepo) GetProject(id int64) (models.Project, error) {
	row := r.database.QueryRow(getProject, id)
	temp := models.Project{}
	err := scanFromRow(row, &temp)
	if err != nil {
		return models.Project{}, err
	}
	return temp, nil
}
