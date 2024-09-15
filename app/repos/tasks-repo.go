package repos

import (
	"database/sql"
	"fmt"
	"go-do-the-thing/app/models"
	"go-do-the-thing/database"
)

type TasksRepo struct {
	database database.DatabaseConnection
}

const TasksRepoName = "items"

func InitTasksRepo(database database.DatabaseConnection) (*TasksRepo, error) {
	_, err := database.Exec(createTasksTable)
	if err != nil {
		return &TasksRepo{}, err
	}
	err = database.AddColumnToTable(TasksRepoName, "tag", "TEXT default '' not null")
	if err != nil {
		return &TasksRepo{}, err
	}
	err = database.AddColumnToTable(TasksRepoName, "complete_date", "TEXT default '' not null")
	if err != nil {
		return &TasksRepo{}, err
	}
	err = database.AddColumnToTable(TasksRepoName, "name", "TEXT default '' not null")
	if err != nil {
		return &TasksRepo{}, err
	}
	return &TasksRepo{database}, nil
}

const (
	createTasksTable = `CREATE TABLE IF NOT EXISTS items (
	[id] INTEGER PRIMARY KEY,
   	[description] TEXT,
	[status] INTEGER DEFAULT 0,
	[assigned_to] TEXT,
    [due_date] TEXT,
    [created_by] TEXT,
    [create_date] TEXT,
	[is_deleted] INTEGER DEFAULT 0
);`
	//getItems           = "SELECT [Id], [description], [status], [assigned_to], [due_date], [created_by], [create_date], [is_deleted], [name], [complete_date] FROM items"
	getItemsNotDeleted = "SELECT [Id], [description], [status], [assigned_to], [due_date], [created_by], [create_date], [is_deleted], [tag], [name], [complete_date] FROM items WHERE is_deleted=0"
	countItems         = "SELECT COUNT(*) FROM items WHERE is_deleted=0"
	getItem            = "SELECT [Id], [description], [status], [assigned_to], [due_date], [created_by], [create_date], [is_deleted], [tag], [name], [complete_date] FROM items WHERE id = %d"
	insertItem         = `INSERT INTO items ([name], [description], [status], [assigned_to], [due_date], [created_by], [create_date], [tag]) VALUES ("%s", "%s", %d, "%s", "%s", "%s", "%s", "%s")`
	updateItemStatus   = `UPDATE items SET [status] = %d, [complete_date] = "%s" WHERE id = %d`
	updateItem         = `UPDATE items SET [name] = "%s", [description] = "%s", [assigned_to] = "%s", [due_date] = "%s", [tag] = "%s" WHERE id = %d`
	deleteItem         = `UPDATE items SET [is_deleted] = 1 WHERE id = %d`
	restoreItem        = `UPDATE items SET [is_deleted] = 0 WHERE id = %d`
)

func scanTaskFromRow(row *sql.Row, item *models.Task) error {
	return row.Scan(
		&item.Id,
		&item.Description,
		&item.Status,
		&item.AssignedTo,
		&item.DueDate,
		&item.CreatedBy,
		&item.CreateDate,
		&item.IsDeleted,
		&item.Tag,
		&item.Name,
		&item.CompleteDate,
	)
}

func scanTaskFromRows(rows *sql.Rows, item *models.Task) error {
	return rows.Scan(
		&item.Id,
		&item.Description,
		&item.Status,
		&item.AssignedTo,
		&item.DueDate,
		&item.CreatedBy,
		&item.CreateDate,
		&item.IsDeleted,
		&item.Tag,
		&item.Name,
		&item.CompleteDate,
	)
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

		err = scanTaskFromRows(rows, &item)
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
	insert := fmt.Sprintf(
		insertItem,
		item.Name,
		item.Description,
		item.Status,
		item.AssignedTo,
		item.DueDate.String(),
		item.CreatedBy,
		item.CreateDate.String(),
		item.Tag,
	)
	result, err := r.database.Exec(insert)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *TasksRepo) UpdateItem(item models.Task) (err error) {
	update := fmt.Sprintf(updateItem, item.Name, item.Description, item.AssignedTo, item.DueDate.String(), item.Tag, item.Id)
	_, err = r.database.Exec(update)
	if err != nil {
		return err
	}
	return nil
}

func (r *TasksRepo) UpdateItemStatus(id int64, completeDate database.SqLiteTime, status int64) (err error) {
	update := fmt.Sprintf(updateItemStatus, status, completeDate.String(), id)
	_, err = r.database.Exec(update)
	if err != nil {
		return err
	}
	return nil
}

func (r *TasksRepo) DeleteItem(id int64) (err error) {
	del := fmt.Sprintf(deleteItem, id)
	_, err = r.database.Exec(del)
	if err != nil {
		return err
	}
	return nil
}

func (r *TasksRepo) RestoreItem(id int64) (err error) {
	res := fmt.Sprintf(restoreItem, id)
	_, err = r.database.Exec(res)
	if err != nil {
		return err
	}
	return nil
}

func (r *TasksRepo) GetItem(id int64) (models.Task, error) {
	get := fmt.Sprintf(getItem, id)
	row := r.database.QueryRow(get)
	temp := models.Task{}
	err := scanTaskFromRow(row, &temp)
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
