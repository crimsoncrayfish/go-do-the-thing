package repos

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
)

type TagsRepo struct {
	logger   slog.Logger
	database database.DatabaseConnection
}

const TagsRepoName = "tags"

// NOTE: READONLY REPO
func InitTagsRepo(database database.DatabaseConnection) (*TagsRepo, error) {
	logger := slog.NewLogger(TagsRepoName)
	_, err := database.Exec(createTagsTable)
	assert.NoError(err, logger, "Failed to create Tags table")
	_, err = database.Exec(seedTagsTable)
	assert.NoError(err, logger, "Failed to seed Tags table")
	return &TagsRepo{
		database: database,
		logger:   logger,
	}, nil
}

const (
	createTagsTable = `CREATE TABLE IF NOT EXISTS tags (
	[id] INTEGER PRIMARY KEY,
   	[name] TEXT DEFAULT '' NOT NULL,
);`
	seedTagsTable = `SOME SQL HERE TO SEED THE TAGS`
	getAllTags    = `SOME SQL HERE TO GET ALL TAGS`
)
