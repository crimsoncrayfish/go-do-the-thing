package tasks_repo

import (
	"database/sql"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/models"
)

type TasksRepo struct {
	database database.DatabaseConnection
}

var repoName = "Tasks Repo"

// NOTE: Depends on: [./projects-repo.go, ./users-repo.go]
func InitRepo(database database.DatabaseConnection) *TasksRepo {
	_, err := database.Exec(createTasksTable)
	assert.NoError(err, repoName, "Failed to create Tasks table")

	return &TasksRepo{database}
}

const createTasksTable = `CREATE TABLE IF NOT EXISTS items (
	[id] INTEGER PRIMARY KEY,
   	[name] TEXT DEFAULT '' NOT NULL,
   	[description] TEXT,
	[assigned_to] INTEGER,
	[project_id] INTEGER,
	[status] INTEGER DEFAULT 0,
	[complete_date] INT DEFAULT 0, 
    [due_date] INT DEFAULT 0,
    [created_by] INTEGER,
    [created_date] INT DEFAULT 0,
    [modified_by] INTEGER,
    [modified_date] INT DEFAULT 0,
	[is_deleted] INTEGER DEFAULT 0,
	FOREIGN KEY (assigned_to) REFERENCES users(id),
	FOREIGN KEY (created_by) REFERENCES users(id),
	FOREIGN KEY (modified_by) REFERENCES users(id),
	FOREIGN KEY (project_id) REFERENCES projects(id)
);`

func scanFromRow(row *sql.Row, item *models.Task) error {
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

func scanFromRows(rows *sql.Rows, item *models.Task) error {
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
		[Id],
		[name], 
		[description],
		[status],	
		[assigned_to],
		[due_date],
		[created_by],
		[created_date],
		[modified_by],
		[modified_date],
		[is_deleted], 
		[project_id],
		[complete_date]
	FROM items
	WHERE [is_deleted] = 0 
	AND [assigned_to] = ?`

func (r *TasksRepo) GetItemsForUser(userId int64) (items []*models.Task, err error) {
	rows, err := r.database.Query(getItemsByAssignedUser, userId)
	if err != nil {
		return nil, errors.New(errors.ErrDBReadFailed, "failed to read items for user: %w", err)
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
		if err != nil {
			err = errors.New(errors.ErrDBGenericError, "failed to close rows: %w", err)
		}
	}(rows)

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
			[name],
			[description],
			[status],
			[assigned_to],
			[due_date],
			[created_by],
			[created_date],
			[modified_by], 
			[modified_date],
			[project_id])
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

func (r *TasksRepo) InsertItem(item models.Task) (id int64, err error) {
	result, err := r.database.Exec(
		insertItem,
		item.Name,
		item.Description,
		item.Status,
		item.AssignedTo,
		item.DueDate,
		item.CreatedBy,
		database.SqLiteNow(),
		item.ModifiedBy,
		database.SqLiteNow(),
		item.Project,
	)
	if err != nil {
		return 0, errors.New(errors.ErrDBReadFailed, "failed to insert task: %w", err)
	}

	last_id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.New(errors.ErrDBReadFailed, "failed to read inserted id: %w", err)
	}

	return last_id, nil
}

const updateItem = `
	UPDATE items
	SET
		[name] = ?,
		[description] = ?,
		[assigned_to] = ?,
		[due_date] = ?,
		[project_id] = ?
	WHERE id = ?`

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
		[status] = ?,
		[complete_date] = ?,
		[modified_by] = ?,
		[modified_date] = ?
	WHERE id = ?`

func (r *TasksRepo) UpdateItemStatus(id int64, completeDate *database.SqLiteTime, status, modifiedBy int64) (err error) {
	_, err = r.database.Exec(updateItemStatus, status, completeDate, modifiedBy, database.SqLiteNow(), id)
	if err != nil {
		return errors.New(errors.ErrDBUpdateFailed, "failed to update the task: %w", err)
	}
	return nil
}

const deleteItem = `UPDATE items SET [is_deleted] = 1, [modified_by] = ?, [modified_date] = ? WHERE id = ?`

func (r *TasksRepo) DeleteItem(id, modifiedBy int64) (err error) {
	_, err = r.database.Exec(deleteItem, modifiedBy, database.SqLiteNow(), id)
	if err != nil {
		return errors.New(errors.ErrDBDeleteFailed, "failed to delete the task: %w", err)
	}
	return nil
}

const restoreItem = `UPDATE items SET [is_deleted] = 0, [modified_by] = ?, [modified_date] = ? WHERE id = ?`

func (r *TasksRepo) RestoreItem(id, modifiedBy int64) (err error) {
	_, err = r.database.Exec(restoreItem, modifiedBy, database.SqLiteNow(), id)
	if err != nil {
		return errors.New(errors.ErrDBUpdateFailed, "failed to restore the task: %w", err)
	}
	return nil
}

const getItem = `
	SELECT 
		[Id], 
		[name],
		[description], 
		[status],
		[assigned_to],
		[due_date],
		[created_by],
		[created_date], 
		[modified_by],
		[modified_date],
		[is_deleted],
		[project_id],
		[complete_date]
	FROM items
	WHERE id = ?`

func (r *TasksRepo) GetItem(id int64) (*models.Task, error) {
	row := r.database.QueryRow(getItem, id)
	temp := &models.Task{}
	err := scanFromRow(row, temp)
	if err != nil {
		return nil, errors.New(errors.ErrDBReadFailed, "failed to get the task: %w", err)
	}
	return temp, nil
}

const countItems = "SELECT COUNT(id) FROM items WHERE [is_deleted]=0 AND [assigned_to]=?"

func (r *TasksRepo) GetItemsCount(userId int64) (int64, error) {
	row := r.database.QueryRow(countItems, userId)
	var temp int64
	err := row.Scan(&temp)
	if err != nil {
		return 0, errors.New(errors.ErrDBReadFailed, "failed to get the task count: %w", err)
	}
	return temp, nil
}
