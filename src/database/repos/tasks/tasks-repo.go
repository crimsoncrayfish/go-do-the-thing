package tasks_repo

import (
	"database/sql"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"
)

type TasksRepo struct {
	database database.DatabaseConnection
}

// NOTE: Depends on: [./projects-repo.go, ./users-repo.go]
func InitRepo(database database.DatabaseConnection) *TasksRepo {
	logger := slog.NewLogger("tasks repo")
	_, err := database.Exec(createTasksTable)
	assert.NoError(err, logger, "Failed to create Tasks table")

	return &TasksRepo{database}
}

const (
	createTasksTable = `CREATE TABLE IF NOT EXISTS items (
	[id] INTEGER PRIMARY KEY,
   	[name] TEXT DEFAULT '' NOT NULL,
   	[description] TEXT,
	[assigned_to] INTEGER,
	[project_id] INTEGER,
	[status] INTEGER DEFAULT 0,
	[complete_date] TEXT DEFAULT '' NOT NULL,	
    [due_date] TEXT,
    [created_by] INTEGER,
    [created_date] TEXT,
    [modified_by] INTEGER,
    [modified_date] TEXT,
	[is_deleted] INTEGER DEFAULT 0,
	FOREIGN KEY (assigned_to) REFERENCES users(id),
	FOREIGN KEY (created_by) REFERENCES users(id),
	FOREIGN KEY (modified_by) REFERENCES users(id),
	FOREIGN KEY (project_id) REFERENCES projects(id)
);`

	getItemsNotDeleted     = "SELECT [Id], [name], [description], [status], [assigned_to], [due_date], [created_by], [created_date], [modified_by], [modified_date], [is_deleted], [project_id], [complete_date]  FROM items WHERE is_deleted=0"
	getItemsByAssignedUser = "SELECT [Id], [name], [description], [status], [assigned_to], [due_date], [created_by], [created_date], [modified_by], [modified_date], [is_deleted], [project_id], [complete_date] FROM items WHERE [is_deleted] = 0 AND [assigned_to] = ?"
	getItem                = "SELECT [Id], [name], [description], [status], [assigned_to], [due_date], [created_by], [created_date], [modified_by], [modified_date], [is_deleted], [project_id], [complete_date] FROM items WHERE id = ?"

	countItems       = "SELECT COUNT(*) FROM items WHERE is_deleted=0"
	insertItem       = `INSERT INTO items ([name], [description], [status], [assigned_to], [due_date], [created_by], [created_date], [modified_by], [modified_date], [project_id]) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	updateItemStatus = `UPDATE items SET [status] = ?, [complete_date] = ?, [modified_by] = ?, [modified_date] = ? WHERE id = ?`
	updateItem       = `UPDATE items SET [name] = ?, [description] = ?, [assigned_to] = ?, [due_date] = ?, [project_id] = ? WHERE id = ?`
	deleteItem       = `UPDATE items SET [is_deleted] = 1, [modified_by] = ?, [modified_date] = ? WHERE id = ?`
	restoreItem      = `UPDATE items SET [is_deleted] = 0 WHERE id = ?`
)

func scanFromRow(row *sql.Row, item *models.Task) error {
	return row.Scan(
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
}

func scanFromRows(rows *sql.Rows, item *models.Task) error {
	return rows.Scan(
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
}

func (r *TasksRepo) GetItemsForUser(userId int64) (items []models.Task, err error) {
	rows, err := r.database.Query(getItemsByAssignedUser, userId)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
	}(rows)

	items = make([]models.Task, 0)
	for rows.Next() {
		item := models.Task{}

		err = scanFromRows(rows, &item)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return items, nil
}
func (r *TasksRepo) GetItems() (items []models.Task, err error) {
	rows, err := r.database.Query(getItemsNotDeleted)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
	}(rows)

	items = make([]models.Task, 0)
	for rows.Next() {
		item := models.Task{}

		err = scanFromRows(rows, &item)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *TasksRepo) InsertItem(item models.Task) (id int64, err error) {
	result, err := r.database.Exec(
		insertItem,
		item.Name,
		item.Description,
		item.Status,
		item.AssignedTo,
		item.DueDate.String(),
		item.CreatedBy,
		database.SqLiteNow().String(),
		item.ModifiedBy,
		database.SqLiteNow().String(),
		item.Project,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *TasksRepo) UpdateItem(item models.Task) (err error) {
	_, err = r.database.Exec(updateItem, item.Name, item.Description, item.AssignedTo, item.DueDate.String(), item.Project, item.Id)
	if err != nil {
		return err
	}
	return nil
}

func (r *TasksRepo) UpdateItemStatus(id int64, completeDate database.SqLiteTime, status, modifiedBy int64) (err error) {
	_, err = r.database.Exec(updateItemStatus, status, completeDate.String(), modifiedBy, database.SqLiteNow().String(), id)
	if err != nil {
		return err
	}
	return nil
}

func (r *TasksRepo) DeleteItem(id, modifiedBy int64, modifiedDate database.SqLiteTime) (err error) {
	_, err = r.database.Exec(deleteItem, modifiedBy, modifiedDate.String(), id)
	if err != nil {
		return err
	}
	return nil
}

func (r *TasksRepo) RestoreItem(id int64) (err error) {
	_, err = r.database.Exec(restoreItem, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *TasksRepo) GetItem(id int64) (models.Task, error) {
	row := r.database.QueryRow(getItem, id)
	temp := models.Task{}
	err := scanFromRow(row, &temp)
	if err != nil {
		return models.Task{}, err
	}
	return temp, nil
}

func (r *TasksRepo) GetItemsCount() (int, error) {
	row := r.database.QueryRow(countItems)
	var temp int
	err := row.Scan(&temp)
	if err != nil {
		return 0, err
	}
	return temp, nil
}
