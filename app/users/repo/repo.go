package usersRepo

import (
	"database/sql"
	"fmt"
	userModel "go-do-the-thing/app/users/model"
	"go-do-the-thing/database"
	"go-do-the-thing/helpers"
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
   	[email] TEXT UNIQUE,
   	[full_name] TEXT DEFAULT "",
    [session_id] TEXT DEFAULT "",
	[session_start_time] TEXT DEFAULT "",
    [password_hash] TEXT DEFAULT "",
	[is_deleted] INTEGER DEFAULT 0,
	[is_admin] INTEGER DEFAULT 0,
	[create_date] TEXT
);`
	getAllUsersNotDeleted = "SELECT [id], [email], [full_name], [session_id], [session_start_time], [is_deleted],[is_admin], [create_date] FROM users WHERE is_deleted=0"
	countUsers            = "SELECT COUNT(*) FROM users WHERE is_deleted=0"
	getUser               = "SELECT [id], [email], [full_name], [session_id], [session_start_time], [is_admin], [is_deleted] FROM users WHERE id = %d"
	getUserByEmail        = `SELECT [id], [email], [full_name], [session_id], [session_start_time], [is_admin], [is_deleted], [create_date] FROM users WHERE email = "%s"`
	insertUser            = `INSERT INTO users ([email], [full_name], [password_hash], [create_date]) VALUES ("%s", "%s", "%s", "%s")`
	updateUserDetails     = `UPDATE users SET [full_name] = "%s" WHERE id = %d`
	updateUserSession     = `UPDATE users SET [session_id] = "%s", [session_start_time] = "%s" WHERE id = %d`
	updateUserPassword    = `UPDATE users SET [password_hash] = "%s" WHERE id = %d`
	updateUserIsAdmin     = `UPDATE users SET [is_admin] = %d WHERE id = %d`
	getUserPassword       = `SELECT [password_hash] FROM [users] WHERE id = %d`
	deleteUser            = `UPDATE users SET [is_deleted] = 1 WHERE id = %d`
	restoreUsers          = `UPDATE users SET [is_deleted] = 0 WHERE id = %d`
	logoutUser            = `UPDATE users SET [session_id] = "", [session_start_time] = "" WHERE id = %d`
)

func ScanItemFromRow(row *sql.Row, user *userModel.User) error {
	return row.Scan(
		&user.Id,
		&user.Email,
		&user.FullName,
		&user.SessionId,
		&user.SessionStartTime,
		&user.IsAdmin,
		&user.IsDeleted,
		&user.CreateDate,
	)
}

func ScanItemFromRows(rows *sql.Rows, user *userModel.User) error {
	return rows.Scan(
		&user.Id,
		&user.Email,
		&user.FullName,
		&user.SessionId,
		&user.SessionStartTime,
		&user.IsDeleted,
		&user.IsAdmin,
		&user.CreateDate,
	)
}

func (r *Repo) Create(user userModel.User) (int64, error) {
	query := fmt.Sprintf(insertUser, user.Email, user.FullName, user.PasswordHash, database.SqLiteNow().String())
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

func (r *Repo) UpdateDetails(user userModel.User) error {
	query := fmt.Sprintf(updateUserDetails, user.FullName, user.Id)
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) UpdatePassword(user userModel.User) error {
	query := fmt.Sprintf(updateUserPassword, user.PasswordHash, user.Id)
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) UpdateSession(user userModel.User) error {
	query := fmt.Sprintf(updateUserSession, user.SessionId, user.SessionStartTime, user.Id)
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) UpdateIsAdmin(user userModel.User) error {
	query := fmt.Sprintf(updateUserIsAdmin, helpers.Btoi(user.IsAdmin), user.Id)
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) Delete(user userModel.User) error {
	query := fmt.Sprintf(deleteUser, user.Id)
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repo) GetUserByEmail(name string) (userModel.User, error) {
	get := fmt.Sprintf(getUserByEmail, name)
	row := r.db.QueryRow(get)
	temp := userModel.User{}
	err := ScanItemFromRow(row, &temp)
	if err != nil {
		// TODO: A couple places rely on this error to determine if a user exitst.
		// What if the scan fails for another reason
		return userModel.User{}, err
	}
	return temp, nil
}

func (r *Repo) GetUserPassword(id int) (string, error) {
	get := fmt.Sprintf(getUserPassword, id)
	row := r.db.QueryRow(get)
	var password string
	err := row.Scan(&password)
	if err != nil {
		// TODO: A couple places rely on this error to determine if a user exitst.
		// What if the scan fails for another reason
		return "", err
	}
	return password, nil
}

func (r *Repo) GetUserById(id int) (userModel.User, error) {
	get := fmt.Sprintf(getUser, id)
	row := r.db.QueryRow(get)
	temp := userModel.User{}
	err := ScanItemFromRow(row, &temp)
	if err != nil {
		return userModel.User{}, err
	}
	return temp, nil
}

func (r *Repo) GetUsers() ([]userModel.User, error) {
	rows, err := r.db.Query(getAllUsersNotDeleted)
	if err != nil {
		return nil, err
	}

	users := make([]userModel.User, 0)
	for rows.Next() {
		user := userModel.User{}

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

func (r *Repo) Logout(userId int64) error {
	query := fmt.Sprintf(logoutUser, userId)
	_, err := r.db.Exec(query)
	return err
}
