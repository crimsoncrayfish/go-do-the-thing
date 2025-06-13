package tags_repo

import (
	"database/sql"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/models"
)

type TagsRepo struct {
	database database.DatabaseConnection
}

var repoName = "Tags Repo"

// NOTE: Depends on: []
func InitRepo(database database.DatabaseConnection) *TagsRepo {
	assert.IsTrue(false, repoName, "not implemented exception")
	//TODO: Cleanup
	//_, err := database.Exec(createTagsTable)
	//assert.NoError(err, repoName, "Failed to create Tags table")
	return &TagsRepo{
		database: database,
	}
}

const (
	getTags = `
		SELECT 
			id,
			name 
		FROM tags 
		WHERE user_id = $1`
	getTag = `
		SELECT 
			id,
			name 
		FROM tags 
		WHERE id = $1`
	insertTag = `
		INSERT INTO tags (
			id,
			name,
			user_id
		) VALUES ($1, $2, $3) 
		ON CONFLICT DO NOTHING`
	deleteTag = `
		DELETE FROM tags 
		WHERE id = $1 
		AND user_id = $2`
)

func scanTagFromRow(row *sql.Row, item *models.Tag) error {
	return row.Scan(
		&item.Id,
		&item.Name,
	)
}

func scanTagFromRows(rows *sql.Rows, item *models.Tag) error {
	return rows.Scan(
		&item.Id,
		&item.Name,
	)
}

func (r *TagsRepo) GetAll(user_id int) (tags []models.Tag, err error) {
	rows, err := r.database.Query(getTags, user_id)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
	}(rows)

	tags = make([]models.Tag, 0)
	for rows.Next() {
		tag := models.Tag{}

		err = scanTagFromRows(rows, &tag)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return tags, nil
}
