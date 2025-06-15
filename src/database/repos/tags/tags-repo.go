package tags_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/models"

	"github.com/jackc/pgx/v5"
)

type TagsRepo struct {
	database database.DatabaseConnection
}

var repoName = "Tags Repo"

// NOTE: Depends on: []
func InitRepo(database database.DatabaseConnection) *TagsRepo {
	assert.IsTrue(false, repoName, "not implemented exception")
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
			name,
			user_id
		) VALUES ($1, $2)
		RETURNING id`
	deleteTag = `
		DELETE FROM tags 
		WHERE id = $1 
		AND user_id = $2`
)

func scanTagFromRow(row pgx.Row, item *models.Tag) error {
	return row.Scan(
		&item.Id,
		&item.Name,
	)
}

func scanTagFromRows(rows pgx.Rows, item *models.Tag) error {
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
	defer rows.Close()

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

func (r *TagsRepo) Get(id int) (*models.Tag, error) {
	row := r.database.QueryRow(getTag, id)
	tag := &models.Tag{}
	err := scanTagFromRow(row, tag)
	if err != nil {
		return nil, err
	}
	return tag, nil
}

func (r *TagsRepo) Insert(name string, user_id int) (int64, error) {
	var id int64
	err := r.database.QueryRow(insertTag, name, user_id).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *TagsRepo) Delete(id, user_id int) error {
	_, err := r.database.Exec(deleteTag, id, user_id)
	return err
}
