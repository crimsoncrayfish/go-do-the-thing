package users

import (
	"database/sql"
	"go-do-the-thing/database"
)

type Repo struct {
	db database.DatabaseConnection
}

func InitRepo(connection database.DatabaseConnection) (Repo, error) {
	//do db migration
	_, err := connection.Exec(createTable)
	if err != nil {
		return Repo{}, err
	}
	return Repo{connection}, nil
}

const (
	createTable = `CREATE TABLE IF NOT EXISTS users (
	[id] INTEGER PRIMARY KEY,
   	[name] TEXT,
	[surname] TEXT,
    [email] TEXT,
    [session_id] TEXT,
	[session_start_time] TEXT,
    [password_hash] TEXT,
	[is_deleted] INTEGER DEFAULT 0
);`
	//getItemsNotDeleted = "SELECT [Id], [description], [status], [assigned_to], [due_date], [created_by], [create_date], [is_deleted], [tag], [name], [complete_date] FROM items WHERE is_deleted=0"
	//countItems         = "SELECT COUNT(*) FROM items WHERE is_deleted=0"
	//getItem            = "SELECT [Id], [description], [status], [assigned_to], [due_date], [created_by], [create_date], [is_deleted], [tag], [name], [complete_date] FROM items WHERE id = %d"
	//insertItem         = `INSERT INTO items ([name], [description], [status], [assigned_to], [due_date], [created_by], [create_date], [tag]) VALUES ("%s", "%s", %d, "%s", "%s", "%s", "%s", "%s")`
	//updateItem         = `UPDATE items SET [name] = "%s", [description] = "%s", [assigned_to] = "%s", [due_date] = "%s", [tag] = "%s" WHERE id = %d`
	//deleteItem         = `UPDATE items SET [is_deleted] = 1 WHERE id = %d`
	//restoreItem        = `UPDATE items SET [is_deleted] = 0 WHERE id = %d`
)

func ScanItemFromRow(row *sql.Row, user *User) error {
	return row.Scan(
		&user.Id,
		&user.Name,
		&user.Surname,
		&user.Email,
		&user.PasswordHash,
		&user.SessionId,
		&user.SessionStartTime,
		&user.IsDeleted,
	)
}

func ScanItemFromRows(rows *sql.Rows, user *User) error {
	return rows.Scan(
		&user.Id,
		&user.Name,
		&user.Surname,
		&user.Email,
		&user.PasswordHash,
		&user.SessionId,
		&user.SessionStartTime,
		&user.IsDeleted,
	)
}
