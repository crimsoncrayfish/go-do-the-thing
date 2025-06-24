package tags_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"

	"github.com/jackc/pgx/v5"
)

type TagsRepo struct {
	database database.DatabaseConnection
	logger   slog.Logger
}

var repoName = "TagsRepo"

// NOTE: Depends on: []
func InitRepo(database database.DatabaseConnection) *TagsRepo {
	logger := slog.NewLogger(repoName)
	return &TagsRepo{database, logger}
}

const (
	getTags   = `SELECT * FROM sp_get_tags($1)`
	getTag    = `SELECT * FROM sp_get_tag($1)`
	insertTag = `SELECT sp_insert_tag($1, $2)`
	deleteTag = `SELECT sp_delete_tag($1, $2)`
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
	r.logger.Debug("GetAll called - sql: %s, params: %d", getTags, user_id)
	rows, err := r.database.Query(getTags, user_id)
	if err != nil {
		r.logger.Error(err, "failed to read tags - sql: %s, params: %d", getTags, user_id)
		return nil, errors.New(errors.ErrDBReadFailed, "failed to read tags: %w", err)
	}
	defer rows.Close()

	tags = make([]models.Tag, 0)
	for rows.Next() {
		tag := models.Tag{}

		err = scanTagFromRows(rows, &tag)
		if err != nil {
			r.logger.Error(err, "failed to scan row in GetAll")
			return nil, err
		}
		tags = append(tags, tag)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error(err, "rows.Err() in GetAll")
		return nil, errors.New(errors.ErrDBGenericError, "some error contained in the rows: %w", err)
	}
	r.logger.Debug("GetAll succeeded - count: %d", len(tags))
	return tags, nil
}

func (r *TagsRepo) Get(id int) (*models.Tag, error) {
	r.logger.Debug("Get called - sql: %s, params: %d", getTag, id)
	row := r.database.QueryRow(getTag, id)
	tag := &models.Tag{}
	err := scanTagFromRow(row, tag)
	if err != nil {
		r.logger.Error(err, "failed to get tag - sql: %s, params: %d", getTag, id)
		return nil, errors.New(errors.ErrDBReadFailed, "failed to get tag: %w", err)
	}
	r.logger.Debug("Get succeeded - id: %d", id)
	return tag, nil
}

func (r *TagsRepo) Insert(name string, user_id int) (int64, error) {
	r.logger.Debug("Insert called - sql: %s, params: %s, %d", insertTag, name, user_id)
	var id int64
	err := r.database.QueryRow(insertTag, name, user_id).Scan(&id)
	if err != nil {
		r.logger.Error(err, "failed to insert tag - sql: %s, params: %s, %d", insertTag, name, user_id)
		return 0, errors.New(errors.ErrDBReadFailed, "failed to insert tag: %w", err)
	}
	r.logger.Info("Tag created successfully - id: %d, name: %s", id, name)
	return id, nil
}

func (r *TagsRepo) Delete(id, user_id int) error {
	r.logger.Debug("Delete called - sql: %s, params: %d, %d", deleteTag, id, user_id)
	_, err := r.database.Exec(deleteTag, id, user_id)
	if err != nil {
		r.logger.Error(err, "failed to delete tag - sql: %s, params: %d, %d", deleteTag, id, user_id)
		return errors.New(errors.ErrDBDeleteFailed, "failed to delete tag: %w", err)
	}
	r.logger.Info("Tag deleted successfully - id: %d", id)
	return nil
}
