package project_tags_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"

	"github.com/jackc/pgx/v5"
)

type ProjectTagsRepo struct {
	database database.DatabaseConnection
	logger   slog.Logger
}

var repoName = "ProjectTagsRepo"

// NOTE: Depends on: [./projects_repo.go]
func InitRepo(database database.DatabaseConnection) *ProjectTagsRepo {
	logger := slog.NewLogger(repoName)
	return &ProjectTagsRepo{database, logger}
}

func scanTagFromRow(row pgx.Row, item *models.Tag) error {
	return row.Scan(
		&item.Id,
		&item.Name,
		&item.Color,
	)
}

func scanTagFromRows(rows pgx.Rows, item *models.Tag) error {
	return rows.Scan(
		&item.Id,
		&item.Name,
		&item.Color,
	)
}

const getTagsByProjectId = `SELECT * FROM sp_get_project_tags($1)`

func (r *ProjectTagsRepo) GetTagsByProjectId(projectId int64) ([]models.Tag, error) {
	r.logger.Debug("GetTagsByProjectId called - sql: %s, params: %d", getTagsByProjectId, projectId)
	rows, err := r.database.Query(getTagsByProjectId, projectId)
	if err != nil {
		r.logger.Error(err, "failed to get tags by project id - sql: %s, params: %d", getTagsByProjectId, projectId)
		return nil, errors.New(errors.ErrDBReadFailed, "failed to get tags by project id: %w", err)
	}
	defer rows.Close()
	tags := make([]models.Tag, 0)
	for rows.Next() {
		tag := models.Tag{}
		err = scanTagFromRows(rows, &tag)
		if err != nil {
			r.logger.Error(err, "failed to scan row in GetTagsByProjectId")
			return nil, err
		}
		tags = append(tags, tag)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error(err, "rows.Err() in GetTagsByProjectId")
		return nil, errors.New(errors.ErrDBGenericError, "some error contained in the rows: %w", err)
	}
	r.logger.Debug("GetTagsByProjectId succeeded - count: %d", len(tags))
	return tags, nil
}

const insertTag = `SELECT sp_insert_tag($1, $2, $3)`

// Create a tag
func (r *ProjectTagsRepo) CreateTag(name, color string, projectId int64) (int64, error) {
	r.logger.Debug("CreateTag called - sql: %s, params: %s, %s, %d", insertTag, name, color, projectId)
	var id int64
	err := r.database.QueryRow(insertTag, name, color, projectId).Scan(&id)
	if err != nil {
		r.logger.Error(err, "failed to insert tag - sql: %s, params: %s, %s, %d", insertTag, name, color, projectId)
		return 0, errors.New(errors.ErrDBReadFailed, "failed to insert tag: %w", err)
	}
	r.logger.Info("Tag created successfully - id: %d, name: %s", id, name)
	return id, nil
}

const updateTag = `SELECT sp_update_tag($1, $2, $3)`

// Update a tag by id
func (r *ProjectTagsRepo) UpdateTag(id int64, name, color string) error {
	r.logger.Debug("UpdateTag called - sql: %s, params: %d, %s, %s", updateTag, id, name, color)
	_, err := r.database.Exec(updateTag, id, name, color)
	if err != nil {
		r.logger.Error(err, "failed to update tag - sql: %s, params: %d, %s, %s", updateTag, id, name, color)
		return errors.New(errors.ErrDBUpdateFailed, "failed to update tag: %w", err)
	}
	r.logger.Info("Tag updated successfully - id: %d", id)
	return nil
}

const getTagById = `Select sp_get_tag($1)`

// Get a tag by id
func (r *ProjectTagsRepo) GetTagById(id int64) (*models.Tag, error) {
	r.logger.Debug("GetTagById called - sql: %s, params: %d", getTagById, id)
	row := r.database.QueryRow(getTagById, id)
	tag := &models.Tag{}
	err := scanTagFromRow(row, tag)
	if err != nil {
		r.logger.Error(err, "failed to get tag by id - sql: %s, params: %d", getTagById, id)
		return nil, errors.New(errors.ErrDBReadFailed, "failed to get tag by id: %w", err)
	}
	r.logger.Debug("GetTagById succeeded - id: %d", id)
	return tag, nil
}

const deleteTag = `Select sp_delete_tag($1)`

// Delete a tag
func (r *ProjectTagsRepo) DeleteTag(id int64) error {
	r.logger.Debug("DeleteTag called - sql: %s, params: %d", deleteTag, id)
	_, err := r.database.Exec(deleteTag, id)
	if err != nil {
		r.logger.Error(err, "failed to delete tag by id - sql: %s, params: %d", deleteTag, id)
		return errors.New(errors.ErrDBReadFailed, "failed to get tag by id: %w", err)
	}
	r.logger.Debug("DeleteTag succeeded - id: %d", id)
	return nil
}
