package tasks_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/models"
	"time"

	"github.com/jackc/pgx/v5"
)

type TasksRepo struct {
	database database.DatabaseConnection
}

var repoName = "Tasks Repo"

// NOTE: Depends on: [./projects-repo.go, ./users-repo.go]
func InitRepo(database database.DatabaseConnection) *TasksRepo {
	//TODO: Cleanup
	//_, err := database.Exec(createTasksTable)
	//assert.NoError(err, repoName, "Failed to create Tasks table")

	return &TasksRepo{database}
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

const getItemsByAssignedUser = `
	SELECT 
		id,
		name, 
		description,
		status,	
		assigned_to,
		due_date,
		created_by,
		created_date,
		modified_by,
		modified_date,
		is_deleted, 
		project_id,
		complete_date
	FROM items
	WHERE is_deleted = false 
	AND assigned_to = $1`

func (r *TasksRepo) GetItemsForUser(userId int64) (items []*models.Task, err error) {
	rows, err := r.database.Query(getItemsByAssignedUser, userId)
	if err != nil {
		return nil, errors.New(errors.ErrDBReadFailed, "failed to read items for user: %w", err)
	}
	defer rows.Close()

	items = make([]*models.Task, 0)
	for rows.Next() {
		item := &models.Task{}

		err = scanFromRows(rows, item)
		if err != nil {
			// NOTE: Already wrapped
			return nil, err
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.New(errors.ErrDBGenericError, "some error contained in the rows: %w", err)
	}
	return items, nil
}

const getItemsByAssignedUserAndProject = `
	SELECT 
		id,
		name, 
		description,
		status,	
		assigned_to,
		due_date,
		created_by,
		created_date,
		modified_by,
		modified_date,
		is_deleted, 
		project_id,
		complete_date
	FROM items
	WHERE is_deleted = false 
	AND assigned_to = $1
	AND project_id = $2`

func (r *TasksRepo) GetItemsForUserAndProject(user_id, project_id int64) (items []*models.Task, err error) {
	rows, err := r.database.Query(getItemsByAssignedUserAndProject, user_id, project_id)
	if err != nil {
		return nil, errors.New(errors.ErrDBReadFailed, "failed to read items for user and project: %w", err)
	}
	defer rows.Close()

	items = make([]*models.Task, 0)
	for rows.Next() {
		item := &models.Task{}

		err = scanFromRows(rows, item)
		if err != nil {
			// NOTE: Already wrapped
			return nil, err
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.New(errors.ErrDBGenericError, "some error contained in the rows: %w", err)
	}
	return items, nil
}

const insertItem = `
	INSERT INTO items 
		(
			name,
			description,
			status,
			assigned_to,
			due_date,
			created_by,
			created_date,
			modified_by, 
			modified_date,
			project_id)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id`

func (r *TasksRepo) InsertItem(item models.Task) (id int64, err error) {
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
		return 0, errors.New(errors.ErrDBReadFailed, "failed to insert task: %w", err)
	}
	return id, nil
}

const updateItem = `
	UPDATE items
	SET
		name = $1,
		description = $2,
		assigned_to = $3,
		due_date = $4,
		project_id = $5
	WHERE id = $6`

func (r *TasksRepo) UpdateItem(item models.Task) (err error) {
	_, err = r.database.Exec(updateItem, item.Name, item.Description, item.AssignedTo, item.DueDate, item.Project, item.Id)
	if err != nil {
		return errors.New(errors.ErrDBUpdateFailed, "failed to update the task: %w", err)
	}

	return nil
}

const updateItemStatus = `
	UPDATE items 
	SET 
		status = $1,
		complete_date = $2,
		modified_by = $3,
		modified_date = $4
	WHERE id = $5`

func (r *TasksRepo) UpdateItemStatus(id int64, completeDate *time.Time, status, modifiedBy int64) (err error) {
	_, err = r.database.Exec(updateItemStatus, status, completeDate, modifiedBy, time.Now(), id)
	if err != nil {
		return errors.New(errors.ErrDBUpdateFailed, "failed to update the task: %w", err)
	}
	return nil
}

const deleteItem = `UPDATE items SET is_deleted = true, modified_by = $1, modified_date = $2 WHERE id = $3`

func (r *TasksRepo) DeleteItem(id, modifiedBy int64) (err error) {
	_, err = r.database.Exec(deleteItem, modifiedBy, time.Now(), id)
	if err != nil {
		return errors.New(errors.ErrDBDeleteFailed, "failed to delete the task: %w", err)
	}
	return nil
}

const restoreItem = `UPDATE items SET is_deleted = false, modified_by = $1, modified_date = $2 WHERE id = $3`

func (r *TasksRepo) RestoreItem(id, modifiedBy int64) (err error) {
	_, err = r.database.Exec(restoreItem, modifiedBy, time.Now(), id)
	if err != nil {
		return errors.New(errors.ErrDBUpdateFailed, "failed to restore the task: %w", err)
	}
	return nil
}

const getItem = `
	SELECT 
		id, 
		name,
		description, 
		status,
		assigned_to,
		due_date,
		created_by,
		created_date, 
		modified_by,
		modified_date,
		is_deleted,
		project_id,
		complete_date
	FROM items
	WHERE id = $1`

func (r *TasksRepo) GetItem(id int64) (*models.Task, error) {
	row := r.database.QueryRow(getItem, id)
	temp := &models.Task{}
	err := scanFromRow(row, temp)
	if err != nil {
		return nil, errors.New(errors.ErrDBReadFailed, "failed to get the task: %w", err)
	}
	return temp, nil
}

const countItems = "SELECT COUNT(id) FROM items WHERE is_deleted=false AND assigned_to=$1"

func (r *TasksRepo) GetItemsCount(userId int64) (int64, error) {
	row := r.database.QueryRow(countItems, userId)
	var temp int64
	err := row.Scan(&temp)
	if err != nil {
		return 0, errors.New(errors.ErrDBReadFailed, "failed to get the task count: %w", err)
	}
	return temp, nil
}
