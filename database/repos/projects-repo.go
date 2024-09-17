package repos

import "go-do-the-thing/database"

type ProjectsRepo struct {
	database database.DatabaseConnection
}

const ProjectsRepoName = "projects"

func InitProjectsRepo(database database.DatabaseConnection) (*ProjectsRepo, error) {
	_, err := database.Exec(createTasksTable)
	if err != nil {
		return &ProjectsRepo{}, err
	}
	return &ProjectsRepo{database}, nil
}

const (
	createProjectsTable = `CREATE TABLE IF NOT EXISTS projects (
	[id] INTEGER PRIMARY KEY,
   	[name] TEXT,
   	[description] TEXT,
	[status] INTEGER,
    [created_by] INTEGER,
    [create_date] TEXT,
	[is_deleted] INTEGER DEFAULT 0
);`
	getProjectsNotDeleted = "SELECT [id], [name], [description], [status], [created_by], [create_date], [is_deleted] FROM projects"
	countProjects         = "SELECT COUNT(*) FROM projects WHERE is_deleted = 0"
	countProjectsByUserId = "SELECT COUNT(*) FROM projects WHERE created_by = ? AND is_deleted=0"
	getProject            = "SELECT [id], [name], [description], [status], [created_by], [create_date], [is_deleted] FROM projects WHERE id = ?"
	insertProject         = "INSERT INTO projects ([name], [description], [status], [created_by], [create_date]) VALUES (?, ?, ?, ?, ?)"
	updateProjectStatus   = "UPDATE projects SET [status] = ? WHERE id = ?"
	updateProject         = "UPDATE projects SET [name] = ?, [description] = ?, [status] = ? WHERE id = ?"
	deleteProject         = "UPDATE projects SET [is_deleted] = 1 WHERE id = ?"
	restoreProject        = "UPDATE projects SET [is_deleted] = 0 WHERE id = ?"
)
