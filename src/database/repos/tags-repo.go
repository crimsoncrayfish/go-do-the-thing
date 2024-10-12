package repos

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
)

type TagsRepo struct {
	database database.DatabaseConnection
}

const tagsRepoName = "tags"

func initTagsRepo(database database.DatabaseConnection) *TagsRepo {
	logger := slog.NewLogger(tagsRepoName)
	_, err := database.Exec(createTagsTable)
	assert.NoError(err, logger, "Failed to create Tags table")
	_, err = database.Exec(seedTagsTable)
	assert.NoError(err, logger, "Failed to create Tags table")
	return &TagsRepo{
		database: database,
	}
}

const (
	createTagsTable = `CREATE TABLE IF NOT EXISTS tags (
	[id] INTEGER,
	[name] INTEGER,
);`
	seedTagsTable = `SOME SQL HERE`
)
