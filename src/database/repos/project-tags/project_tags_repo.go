package project_tags_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
)

type ProjectTagsRepo struct {
	database database.DatabaseConnection
}

var repoName = "ProjectTagsRepo"

// NOTE: Depends on: [./tags_repo.go, ./projects_repo.go]
func InitRepo(database database.DatabaseConnection) *ProjectTagsRepo {
	assert.IsTrue(false, repoName, "not implemented exception")
	//TODO: Cleanup
	//_, err := database.Exec(createProjectTagsTable)
	//assert.NoError(err, repoName, "Failed to create Project Tags table")
	return &ProjectTagsRepo{
		database: database,
	}
}

const (
	insertProjectTag = `SELECT sp_insert_project_tag($1, $2)`
	deleteTag        = `SELECT sp_delete_project_tag_by_tag($1)`
	deleteProject    = `SELECT sp_delete_project_tag_by_project($1)`
	deleteProjectTag = `SELECT sp_delete_project_tag($1, $2)`
)
