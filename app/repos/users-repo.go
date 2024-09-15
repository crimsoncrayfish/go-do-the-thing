package repos

import (
	"database/sql"
	"fmt"
	"go-do-the-thing/app/models"
	"go-do-the-thing/database"
	"go-do-the-thing/helpers"
)

type UsersRepo struct {
	db database.DatabaseConnection
}

func InitUsersRepo(connection database.DatabaseConnection) (*UsersRepo, error) {
	//do db migration
	_, err := connection.Exec(createTable)
	if err != nil {
		return &UsersRepo{}, err
	}
	return &UsersRepo{connection}, nil
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

func scanUserFromRow(row *sql.Row, user *models.User) error {
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

func scanUsersFromRows(rows *sql.Rows, user *models.User) error {
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

func (r *UsersRepo) Create(user models.User) (int64, error) {
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

func (r *UsersRepo) UpdateDetails(user models.User) error {
	query := fmt.Sprintf(updateUserDetails, user.FullName, user.Id)
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (r *UsersRepo) UpdatePassword(user models.User) error {
	query := fmt.Sprintf(updateUserPassword, user.PasswordHash, user.Id)
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (r *UsersRepo) UpdateSession(user models.User) error {
	query := fmt.Sprintf(updateUserSession, user.SessionId, user.SessionStartTime, user.Id)
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (r *UsersRepo) UpdateIsAdmin(user models.User) error {
	query := fmt.Sprintf(updateUserIsAdmin, helpers.Btoi(user.IsAdmin), user.Id)
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (r *UsersRepo) Delete(user models.User) error {
	query := fmt.Sprintf(deleteUser, user.Id)
	_, err := r.db.Exec(query)
	if err != nil {
		return err
	}
	return nil
}

func (r *UsersRepo) GetUserByEmail(name string) (models.User, error) {
	get := fmt.Sprintf(getUserByEmail, name)
	row := r.db.QueryRow(get)

	temp := models.User{}
	err := scanUserFromRow(row, &temp)
	if err != nil {
		// TODO: A couple places rely on this error to determine if a user exitst.
		// What if the scan fails for another reason
		return models.User{}, err
	}
	return temp, nil
}

func (r *UsersRepo) GetUserPassword(id int) (string, error) {
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

func (r *UsersRepo) GetUserById(id int) (models.User, error) {
	get := fmt.Sprintf(getUser, id)
	row := r.db.QueryRow(get)
	temp := models.User{}
	err := scanUserFromRow(row, &temp)
	if err != nil {
		return models.User{}, err
	}
	return temp, nil
}

func (r *UsersRepo) GetUsers() ([]models.User, error) {
	rows, err := r.db.Query(getAllUsersNotDeleted)
	if err != nil {
		return nil, err
	}

	users := make([]models.User, 0)
	for rows.Next() {
		user := models.User{}

		err := scanUsersFromRows(rows, &user)
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

func (r *UsersRepo) Logout(userId int64) error {
	query := fmt.Sprintf(logoutUser, userId)
	_, err := r.db.Exec(query)
	return err
}
