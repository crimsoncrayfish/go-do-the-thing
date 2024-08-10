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

func Init(database database.DatabaseConnection) (Repo, error) {
	_, err := database.Exec(createTable)
	if err != nil {
		return Repo{}, err
	}
	err = database.AddColumnToTable(RepoName, "tag", "TEXT default 'a' not null")
	if err != nil {
		return Repo{}, err
	}
	return Repo{database}, nil
}

func (r *Repo) GetItems() (items []Item, err error) {
	rows, err := r.database.Query(getItemsNotDeleted)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
	}(rows)

	items = make([]Item, 0)
	for rows.Next() {
		item := Item{}

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

func (r *Repo) InsertItem(item Item) (id int64, err error) {
	insert := fmt.Sprintf(
		insertItem,
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

func (r *Repo) UpdateItemStatus(id int64, status int64) (err error) {
	update := fmt.Sprintf(updateItemStatus, status, id)
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

func (r *Repo) GetItem(id int64) (Item, error) {
	get := fmt.Sprintf(getItem, id)
	row := r.database.QueryRow(get)
	temp := Item{}
	err := ScanItemFromRow(row, &temp)
	if err != nil {
		return Item{}, err
	}
	return temp, nil
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
	[is_deleted] INTEGER
);`
	getItems           = "SELECT [Id], [description], [status], [assigned_to], [due_date], [created_by], [create_date], [is_deleted] FROM items"
	getItemsNotDeleted = "SELECT [Id], [description], [status], [assigned_to], [due_date], [created_by], [create_date], [is_deleted], [tag] FROM items WHERE is_deleted=0"
	getItem            = "SELECT [Id], [description], [status], [assigned_to], [due_date], [created_by], [create_date], [is_deleted], [tag] FROM items WHERE id = %d"
	insertItem         = `INSERT INTO items ([description], [status], [assigned_to], [due_date], [created_by], [create_date], [tag]) VALUES ("%s", %d, "%s", "%s", "%s", "%s", "%s")`
	updateItemStatus   = `UPDATE items SET [status] = %d WHERE id = %d`
	deleteItem         = `UPDATE items SET [is_deleted] = 1 WHERE id = %d`
	restoreItem        = `UPDATE items SET [is_deleted] = 0 WHERE id = %d`
)

func ScanItemFromRow(row *sql.Row, item *Item) error {
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
	)
}

func ScanItemFromRows(rows *sql.Rows, item *Item) error {
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
	)
}
