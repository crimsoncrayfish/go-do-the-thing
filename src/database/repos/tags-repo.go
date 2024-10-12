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
	return &TagsRepo{
		database: database,
	}
}

const (
	createTagsTable = `CREATE TABLE IF NOT EXISTS tags (
	[id] INTEGER,
	[name] INTEGER,
	[user_id] INTEGER
);`
	insertTag = `INSERT OR IGNORE INTO tags(id, name, user_id) VALUES(?, ?, ?)`
	getTags   = `SELECT id, name FROM tags WHERE [user_id] = ?`
	deleteTag = `DELETE FROM tags WHERE [id] = ?, [user_id] = ?`
)
