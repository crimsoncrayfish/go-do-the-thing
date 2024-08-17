package users

import (
	"database/sql"
	"fmt"
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
	[is_deleted] INTEGER DEFAULT 0,
	[is_admin] INTEGER DEFAULT 0
);`
	getAllUsersNotDeleted = "SELECT [id], [name], [surname], [email], [is_admin] FROM users WHERE is_deleted=0"
	countUsers            = "SELECT COUNT(*) FROM users WHERE is_deleted=0"
	getUser               = "SELECT [id], [name], [surname], [email], [session_id], [session_start_time], [is_admin], [is_deleted] FROM users WHERE id = %d"
	insertUser            = `INSERT INTO users ([name], [surname], [email], [password_hash]) VALUES ("%s", "%s", "%s", "%s")`
	updateUserDetails     = `UPDATE users SET [name] = "%s", [surname] = "%s", [email] = "%s" WHERE id = %d`
	updateUserSession     = `UPDATE users SET [session_id] = "%s", [sessio_start_time] = "%s" WHERE id = %d`
	updateUserPassword    = `UPDATE users SET [password_hash] = "%s" WHERE id = %d`
	updateUserIsAdmin     = `UPDATE users SET [is_admin] = %d WHERE id = %d`
	deleteUser            = `UPDATE users SET [is_deleted] = 1 WHERE id = %d`
	restoreUsers          = `UPDATE users SET [is_deleted] = 0 WHERE id = %d`
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

func (r *Repo) Create(user User) (int64, error) {
	query := fmt.Sprintf(insertUser, user.Name, user.Surname, user.Email, user.PasswordHash)
	result, err := r.db.Exec(query)
	if err != nil {
		return 0, err
	}
	insertedId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return insertedId, nil
}

func (r *Repo) UpdateDetails(user User) error {
	query := fmt.Sprintf(updateUserDetails, user.Name, user.Surname, user.Email, user.Id)
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) UpdatePassword(user User) error {
	query := fmt.Sprintf(updateUserPassword, user.PasswordHash, user.Id)
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) UpdateSession(user User) error {
	query := fmt.Sprintf(updateUserSession, user.SessionId, user.SessionStartTime, user.Id)
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) UpdateIsAdmin(user User) error {
	query := fmt.Sprintf(updateUserIsAdmin, user.IsAdmin, user.Id)
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) Delete(user User) error {
	query := fmt.Sprintf(deleteUser, user.Id)
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) GetUserById(id int) (User, error) {
	get := fmt.Sprintf(getUser, id)
	row := r.db.QueryRow(get)
	temp := User{}
	err := ScanItemFromRow(row, &temp)
	if err != nil {
		return User{}, err
	}
	return temp, nil
}

func (r *Repo) GetUsers() ([]User, error) {
	rows, err := r.db.Query(getAllUsersNotDeleted)
	if err != nil {
		return nil, err
	}

	users := make([]User, 0)
	for rows.Next() {
		user := User{}

		err := ScanItemFromRows(rows, &user)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return users, nil
}
