package project_tags_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/models"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

type ProjectTagsRepo struct {
	database database.DatabaseConnection
	logger   *slog.Logger
}

// NOTE: Depends on: [./projects_repo.go]
func InitRepo(database database.DatabaseConnection) *ProjectTagsRepo {
	logger := slog.With("source", "ProjectTagsRepo")
	return &ProjectTagsRepo{
		database: database,
		logger:   logger,
	}
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
	rows, err := r.database.Query(getTagsByProjectId, projectId)
	if err != nil {
		r.logger.Error("failed to get tags by project id", "error", err, "sql", getTagsByProjectId, "project_id", projectId)
		return nil, errors.New(errors.ErrDBReadFailed, "failed to get tags by project id: %w", err)
	}
	defer rows.Close()
	tags := make([]models.Tag, 0)
	for rows.Next() {
		tag := models.Tag{}
		err = scanTagFromRows(rows, &tag)
		if err != nil {
			r.logger.Error("failed to scan row in GetTagsByProjectId", "error", err)
			return nil, err
		}
		tags = append(tags, tag)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error("rows.Err() in GetTagsByProjectId", "error", err)
		return nil, errors.New(errors.ErrDBGenericError, "some error contained in the rows: %w", err)
	}
	r.logger.Debug("GetTagsByProjectId succeeded", "count", len(tags))
	return tags, nil
}

const insertTag = `SELECT sp_insert_tag($1, $2, $3, $4)`

func (r *ProjectTagsRepo) CreateTag(name string, color models.TagColor, current_user_id, project_id int64) (int64, error) {
	var id int64
	err := r.database.QueryRow(insertTag, name, color, current_user_id, project_id).Scan(&id)
	if err != nil {
		r.logger.Error("failed to insert tag", "error", err, "sql", insertTag, "name", name, "color", color, "project_id", project_id)
		return 0, errors.New(errors.ErrDBReadFailed, "failed to insert tag: %w", err)
	}
	r.logger.Info("Tag created successfully", "id", id, "name", name)
	return id, nil
}

const updateTag = `SELECT sp_update_tag($1, $2, $3)`

func (r *ProjectTagsRepo) UpdateTag(id int64, name string, color models.TagColor) error {
	_, err := r.database.Exec(updateTag, id, name, color)
	if err != nil {
		r.logger.Error("failed to update tag", "error", err, "sql", updateTag, "id", id, "name", name, "color", color)
		return errors.New(errors.ErrDBUpdateFailed, "failed to update tag: %w", err)
	}
	r.logger.Info("Tag updated successfully", "id", id)
	return nil
}

const getTagById = `Select * FROM sp_get_tag($1, $2)`

func (r *ProjectTagsRepo) GetTagById(id, project_id int64) (*models.Tag, error) {
	row := r.database.QueryRow(getTagById, id, project_id)
	tag := &models.Tag{}
	err := scanTagFromRow(row, tag)
	if err != nil {
		r.logger.Error("failed to get tag by id", "error", err, "sql", getTagById, "id", id)
		return nil, errors.New(errors.ErrDBReadFailed, "failed to get tag by id: %w", err)
	}
	r.logger.Debug("GetTagById succeeded", "id", id)
	return tag, nil
}

const deleteTag = `Select sp_delete_tag($1)`

func (r *ProjectTagsRepo) DeleteTag(id int64) error {
	_, err := r.database.Exec(deleteTag, id)
	if err != nil {
		r.logger.Error("failed to delete tag by id", "error", err, "sql", deleteTag, "id", id)
		return errors.New(errors.ErrDBReadFailed, "failed to get tag by id: %w", err)
	}
	r.logger.Debug("DeleteTag succeeded", "id", id)
	return nil
}
