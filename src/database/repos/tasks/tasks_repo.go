package tasks_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"
	"time"

	"github.com/jackc/pgx/v5"
)

type TasksRepo struct {
	database database.DatabaseConnection
	logger   slog.Logger
}

var repoName = "TasksRepo"

// NOTE: Depends on: [./projects_repo.go, ./users_repo.go]
func InitRepo(database database.DatabaseConnection) *TasksRepo {
	logger := slog.NewLogger(repoName)
	//TODO: Cleanup
	//_, err := database.Exec(createTasksTable)
	//assert.NoError(err, repoName, "Failed to create Tasks table")

	return &TasksRepo{database, logger}
}

func scanFromRow(row pgx.Row, item *models.Task) error {
	err := row.Scan(
		&item.Id,
		&item.Name,
		&item.Description,
		&item.Status,
		&item.AssignedTo,
		&item.DueDate,
		&item.CreatedBy,
		&item.CreatedDate,
		&item.ModifiedBy,
		&item.ModifiedDate,
		&item.IsDeleted,
		&item.Project,
		&item.CompleteDate,
	)
	if err != nil {
		return errors.New(errors.ErrDBGenericError, "failed to scan the row: %w", err)
	}
	return nil
}

func scanFromRows(rows pgx.Rows, item *models.Task) error {
	err := rows.Scan(
		&item.Id,
		&item.Name,
		&item.Description,
		&item.Status,
		&item.AssignedTo,
		&item.DueDate,
		&item.CreatedBy,
		&item.CreatedDate,
		&item.ModifiedBy,
		&item.ModifiedDate,
		&item.IsDeleted,
		&item.Project,
		&item.CompleteDate,
	)
	if err != nil {
		return errors.New(errors.ErrDBGenericError, "failed to scan the rows: %w", err)
	}
	return nil
}

const getItemsByAssignedUser = `SELECT * FROM sp_get_items_by_user($1)`

func (r *TasksRepo) GetItemsForUser(userId int64) (items []*models.Task, err error) {
	r.logger.Debug("GetItemsForUser called - sql: %s, params: %v", getItemsByAssignedUser, userId)
	rows, err := r.database.Query(getItemsByAssignedUser, userId)
	if err != nil {
		r.logger.Error(err, "failed to read items for user - sql: %s, params: %v", getItemsByAssignedUser, userId)
		return nil, errors.New(errors.ErrDBReadFailed, "failed to read items for user: %w", err)
	}
	defer rows.Close()

	items = make([]*models.Task, 0)
	for rows.Next() {
		item := &models.Task{}

		err = scanFromRows(rows, item)
		if err != nil {
			// NOTE: Already wrapped
			r.logger.Error(err, "failed to scan row in GetItemsForUser - params: %v", userId)
			return nil, err
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error(err, "rows.Err() in GetItemsForUser - params: %v", userId)
		return nil, errors.New(errors.ErrDBGenericError, "some error contained in the rows: %w", err)
	}
	r.logger.Debug("GetItemsForUser succeeded - count: %d, params: %v", len(items), userId)
	return items, nil
}

const getItemsByAssignedUserAndProject = `SELECT * FROM sp_get_items_by_user_and_project($1, $2)`

func (r *TasksRepo) GetItemsForUserAndProject(user_id, project_id int64) (items []*models.Task, err error) {
	r.logger.Debug("GetItemsForUserAndProject called - sql: %s, params: %v", getItemsByAssignedUserAndProject, []int64{user_id, project_id})
	rows, err := r.database.Query(getItemsByAssignedUserAndProject, user_id, project_id)
	if err != nil {
		r.logger.Error(err, "failed to read items for user and project - sql: %s, params: %v", getItemsByAssignedUserAndProject, []int64{user_id, project_id})
		return nil, errors.New(errors.ErrDBReadFailed, "failed to read items for user and project: %w", err)
	}
	defer rows.Close()

	items = make([]*models.Task, 0)
	for rows.Next() {
		item := &models.Task{}

		err = scanFromRows(rows, item)
		if err != nil {
			r.logger.Error(err, "failed to scan row in GetItemsForUserAndProject - params: %v", []int64{user_id, project_id})
			return nil, err
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error(err, "rows.Err() in GetItemsForUserAndProject - params: %v", []int64{user_id, project_id})
		return nil, errors.New(errors.ErrDBGenericError, "some error contained in the rows: %w", err)
	}
	r.logger.Debug("GetItemsForUserAndProject succeeded - count: %d, params: %v", len(items), []int64{user_id, project_id})
	return items, nil
}

const insertItem = `SELECT sp_insert_item($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

func (r *TasksRepo) InsertItem(item models.Task) (id int64, err error) {
	r.logger.Debug("InsertItem called - sql: %s, params: %+v", insertItem, item)
	err = r.database.QueryRow(
		insertItem,
		item.Name,
		item.Description,
		item.Status,
		item.AssignedTo,
		item.DueDate,
		item.CreatedBy,
		time.Now(),
		item.ModifiedBy,
		time.Now(),
		item.Project,
	).Scan(&id)
	if err != nil {
		r.logger.Error(err, "failed to insert task - sql: %s, params: %+v", insertItem, item)
		return 0, errors.New(errors.ErrDBReadFailed, "failed to insert task: %w", err)
	}
	r.logger.Info("Task created successfully - id: %d, name: %s, project: %d", id, item.Name, item.Project)
	return id, nil
}

const updateItem = `SELECT sp_update_item($1, $2, $3, $4, $5, $6)`

func (r *TasksRepo) UpdateItem(item models.Task) (err error) {
	r.logger.Debug("UpdateItem called - sql: %s, params: %+v", updateItem, item)
	_, err = r.database.Exec(updateItem, item.Id, item.Name, item.Description, item.AssignedTo, item.DueDate, item.Project)
	if err != nil {
		r.logger.Error(err, "failed to update the task - sql: %s, params: %+v", updateItem, item)
		return errors.New(errors.ErrDBUpdateFailed, "failed to update the task: %w", err)
	}
	r.logger.Info("Task updated successfully - id: %d, name: %s", item.Id, item.Name)
	return nil
}

const updateItemStatus = `SELECT sp_update_item_status($1, $2, $3, $4, $5)`

func (r *TasksRepo) UpdateItemStatus(id int64, completeDate *time.Time, status, modifiedBy int64) (err error) {
	r.logger.Debug("UpdateItemStatus called - sql: %s, params: %v", updateItemStatus, []interface{}{id, status, completeDate, modifiedBy, time.Now()})
	_, err = r.database.Exec(updateItemStatus, id, status, completeDate, modifiedBy, time.Now())
	if err != nil {
		r.logger.Error(err, "failed to update the task status - sql: %s, params: %v", updateItemStatus, []interface{}{id, status, completeDate, modifiedBy, time.Now()})
		return errors.New(errors.ErrDBUpdateFailed, "failed to update the task: %w", err)
	}
	r.logger.Info("Task status updated successfully - id: %d, status: %d", id, status)
	return nil
}

const deleteItem = `SELECT sp_delete_item($1, $2, $3)`

func (r *TasksRepo) DeleteItem(id, modifiedBy int64) (err error) {
	r.logger.Debug("DeleteItem called - sql: %s, params: %v", deleteItem, []interface{}{id, modifiedBy, time.Now()})
	_, err = r.database.Exec(deleteItem, id, modifiedBy, time.Now())
	if err != nil {
		r.logger.Error(err, "failed to delete the task - sql: %s, params: %v", deleteItem, []interface{}{id, modifiedBy, time.Now()})
		return errors.New(errors.ErrDBDeleteFailed, "failed to delete the task: %w", err)
	}
	r.logger.Info("Task deleted successfully - id: %d", id)
	return nil
}

const restoreItem = `SELECT sp_restore_item($1, $2, $3)`

func (r *TasksRepo) RestoreItem(id, modifiedBy int64) (err error) {
	r.logger.Debug("RestoreItem called - sql: %s, params: %v", restoreItem, []interface{}{id, modifiedBy, time.Now()})
	_, err = r.database.Exec(restoreItem, id, modifiedBy, time.Now())
	if err != nil {
		r.logger.Error(err, "failed to restore the task - sql: %s, params: %v", restoreItem, []interface{}{id, modifiedBy, time.Now()})
		return errors.New(errors.ErrDBUpdateFailed, "failed to restore the task: %w", err)
	}
	r.logger.Info("Task restored successfully - id: %d", id)
	return nil
}

const getItem = `SELECT * FROM sp_get_item($1)`

func (r *TasksRepo) GetItem(id int64) (*models.Task, error) {
	r.logger.Debug("GetItem called - sql: %s, params: %v", getItem, id)
	row := r.database.QueryRow(getItem, id)
	temp := &models.Task{}
	err := scanFromRow(row, temp)
	if err != nil {
		r.logger.Error(err, "failed to get the task - sql: %s, params: %v", getItem, id)
		return nil, errors.New(errors.ErrDBReadFailed, "failed to get the task: %w", err)
	}
	r.logger.Debug("GetItem succeeded - id: %d", id)
	return temp, nil
}

const getItemsCount = `SELECT sp_get_items_count($1)`

func (r *TasksRepo) GetItemsCount(userId int64) (int64, error) {
	r.logger.Debug("GetItemsCount called - sql: %s, params: %v", getItemsCount, userId)
	row := r.database.QueryRow(getItemsCount, userId)
	var count int64
	err := row.Scan(&count)
	if err != nil {
		r.logger.Error(err, "failed to get items count - sql: %s, params: %v", getItemsCount, userId)
		return 0, errors.New(errors.ErrDBReadFailed, "failed to get items count: %w", err)
	}
	r.logger.Debug("GetItemsCount succeeded - count: %d, params: %v", count, userId)
	return count, nil
}

// Returns (completed, total, error)
func (r *TasksRepo) GetProjectTaskCompletion(projectId int64) (int64, int64, error) {
	const totalQuery = `SELECT COUNT(*) FROM items WHERE project_id = $1 AND is_deleted = FALSE`
	const completedQuery = `SELECT COUNT(*) FROM items WHERE project_id = $1 AND is_deleted = FALSE AND status = 1`
	var total, completed int64
	err := r.database.QueryRow(totalQuery, projectId).Scan(&total)
	if err != nil {
		r.logger.Error(err, "failed to count total tasks for project", projectId)
		return 0, 0, err
	}
	err = r.database.QueryRow(completedQuery, projectId).Scan(&completed)
	if err != nil {
		r.logger.Error(err, "failed to count completed tasks for project", projectId)
		return completed, total, err
	}
	return completed, total, nil
}
