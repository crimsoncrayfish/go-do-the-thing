package todo

import (
	"database/sql"
	"fmt"
	"go-do-the-thing/database"
)

type Repo struct {
	database database.DatabaseConnection
}

const RepoName = "items"

func InitRepo(database database.DatabaseConnection) (Repo, error) {
	_, err := database.Exec(createTable)
	if err != nil {
		return Repo{}, err
	}
	err = database.AddColumnToTable(RepoName, "tag", "TEXT default '' not null")
	if err != nil {
		return Repo{}, err
	}
	err = database.AddColumnToTable(RepoName, "complete_date", "TEXT default '' not null")
	if err != nil {
		return Repo{}, err
	}
	err = database.AddColumnToTable(RepoName, "name", "TEXT default '' not null")
	if err != nil {
		return Repo{}, err
	}
	return Repo{database}, nil
}

const (
	createTable = `CREATE TABLE IF NOT EXISTS items (
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

func ScanItemFromRow(row *sql.Row, item *Task) error {
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

func ScanItemFromRows(rows *sql.Rows, item *Task) error {
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

func (r *Repo) GetItems() (items []Task, err error) {
	rows, err := r.database.Query(getItemsNotDeleted)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
	}(rows)

	items = make([]Task, 0)
	for rows.Next() {
		item := Task{}

		err = ScanItemFromRows(rows, &item)
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

func (r *Repo) InsertItem(item Task) (id int64, err error) {
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

func (r *Repo) UpdateItem(item Task) (err error) {
	update := fmt.Sprintf(updateItem, item.Name, item.Description, item.AssignedTo, item.DueDate.String(), item.Tag, item.Id)
	_, err = r.database.Exec(update)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) UpdateItemStatus(id int64, completeDate database.SqLiteTime, status int64) (err error) {
	update := fmt.Sprintf(updateItemStatus, status, completeDate.String(), id)
	_, err = r.database.Exec(update)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) DeleteItem(id int64) (err error) {
	del := fmt.Sprintf(deleteItem, id)
	_, err = r.database.Exec(del)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) RestoreItem(id int64) (err error) {
	res := fmt.Sprintf(restoreItem, id)
	_, err = r.database.Exec(res)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) GetItem(id int64) (Task, error) {
	get := fmt.Sprintf(getItem, id)
	row := r.database.QueryRow(get)
	temp := Task{}
	err := ScanItemFromRow(row, &temp)
	if err != nil {
		return Task{}, err
	}
	return temp, nil
}

func (r *Repo) GetItemsCount() (int, error) {
	row := r.database.QueryRow(countItems)
	var temp int
	err := row.Scan(&temp)
	if err != nil {
		return 0, err
	}
	return temp, nil
}