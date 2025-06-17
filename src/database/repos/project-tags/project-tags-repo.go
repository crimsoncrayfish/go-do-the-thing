package project_tags_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
)

type ProjectTagsRepo struct {
	database database.DatabaseConnection
}

var repoName = "ProjectTagsRepo"

// NOTE: Depends on: [./tags-repo.go, ./projects-repo.go]
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
	insertProjectTag = `INSERT INTO project_tags (project_id, tag_id) VALUES (?, ?)`
	deleteTag        = `DELETE FROM project_tags WHERE [tag_id] = ?`
	deleteProject    = `DELETE FROM project_tags WHERE [project_id] = ?`
	deleteProjectTag = `DELETE FROM project_tags WHERE [tag_id] = ? AND [project_id] = ?`
)
