// Package tasktagsrepo The repository layer for task tags
package tasktagsrepo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/logger"
	"go-do-the-thing/src/models"

	"github.com/jackc/pgx/v5"
)

type TaskTagsRepo struct {
	database *database.DatabaseConnection
	logger   *logger.Logger
}

func InitRepo(database *database.DatabaseConnection) *TaskTagsRepo {
	// NOTE: Depends on: [./tags_repo.go, ./tasks_repo.go, ./users_repo.go]
	log := logger.NewLogger("TasksRepo")

	return &TaskTagsRepo{
		logger:   log,
		database: database,
	}
}

func scanFromRows(rows pgx.Rows, task *models.Tag) error {
	err := rows.Scan(
		&task.Id,
		&task.Name,
		&task.Color,
	)
	if err != nil {
		return errors.NewDbGenericError("failed to scan task tag rows", err, "", nil)
	}
	return nil
}

const getTagsByTaskID = `SELECT * FROM sp_get_task_tags($1)`

func (r *TaskTagsRepo) GetTagsByTaskID(taskID int64) ([]models.Tag, error) {
	rows, err := r.database.Query(getTagsByTaskID, taskID)
	if err != nil {
		r.logger.Error("failed to get tags by task id", "error", err, "sql", getTagsByTaskID, "task_id", taskID)
		return nil, errors.NewDbReadFailed("tags", err, "by task id", nil)
	}
	defer rows.Close()
	tags := make([]models.Tag, 0)
	for rows.Next() {
		tag := models.Tag{}
		err = scanFromRows(rows, &tag)
		if err != nil {
			r.logger.Error("failed to scan row in GetTagsByTaskId", "error", err)
			return nil, err
		}
		tags = append(tags, tag)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error("rows.Err() in GetTagsByTaskId", "error", err)
		return nil, errors.NewDbGenericError("rows error in GetTagsByTaskId", err, "", nil)
	}
	r.logger.Debug("GetTagsByTaskId succeeded", "count", len(tags))
	return tags, nil
}

const setTags = `SELECT sp_set_task_tags($1, $2, $3)`

func (r *TaskTagsRepo) SetTaskTags(projectID, taskID int64, tagIDs []int64) error {
	_, err := r.database.Exec(setTags, projectID, taskID, tagIDs)
	if err != nil {
		r.logger.Error("failed to update task tags", "error", err, "sql", setTags, "task", taskID)
		return errors.NewDbUpdateFailed("taskTags", err, "", nil)
	}
	r.logger.Info("Task Tags updated successfully")
	return nil
}
