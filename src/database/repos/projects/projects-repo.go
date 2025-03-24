package projects_repo

import (
	"database/sql"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/models"
)

type ProjectsRepo struct {
	database database.DatabaseConnection
}

var repoName = assert.Source{"ProjectsRepo"}

// NOTE: Depends on: [../project-users/project-users-repo.go, ../users/users-repo.go]
func InitRepo(database database.DatabaseConnection) *ProjectsRepo {
	_, err := database.Exec(createProjectsTable)
	assert.NoError(err, repoName, "Failed to create Projects table")

	return &ProjectsRepo{
		database: database,
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
		[id],[name],[description],[owner],[start_date],[due_date],[created_by],[create_date],[modified_by],[modified_date],[is_complete],[is_deleted] 
	FROM projects 
	JOIN project_users
		ON projects.id = project_users.project_id
	WHERE project_users.user_id = ?`

func (r *ProjectsRepo) GetProjects(user_id int) ([]models.Project, error) {
	rows, err := r.database.Query(getProjectsByUser, user_id)
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

const getProject = `
	SELECT 
		[id],[name],[description],[owner],[start_date],[due_date],[created_by],[create_date],[modified_by],[modified_date],[is_complete],[is_deleted]
	FROM projects
	JOIN project_users
		ON projects.id = project_users.project_id
	WHERE 
		project.id = ? AND
		project_user.user_id = ?`

func (r *ProjectsRepo) GetProject(projectId, userId int64) (models.Project, error) {
	row := r.database.QueryRow(getProject, projectId, userId)
	temp := models.Project{}
	err := scanFromRow(row, &temp)
	if err != nil {
		return models.Project{}, err
	}
	return temp, nil
}

const deleteProject = `
	UPDATE projects 
	SET [is_deleted] = 1, [modified_by] = ?, [modified_date] = ?
	WHERE id = ?`

func (r *ProjectsRepo) DeleteProject(id, currentUser int64) error {
	_, err := r.database.Exec(deleteProject, id, currentUser, database.SqLiteNow())
	return err
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
		[modified_date] = ?,
		[is_complete] = ?,
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
		project.IsComplete,
	)
	return err
}

const insertProject = `
	INSERT INTO projects 
	([name],[description],[owner],[start_date],[due_date],[created_by],[create_date],[is_complete])
	VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`

func (r *ProjectsRepo) InsertProject(currentUser int64, project models.Project) (err error) {
	_, err = r.database.Exec(updateProject,
		project.Name,
		project.Description,
		project.Owner,
		project.StartDate,
		project.DueDate,
		currentUser,
		database.SqLiteNow(),
		project.IsComplete,
	)
	return err
}
