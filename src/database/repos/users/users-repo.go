package users_repo

import (
	"database/sql"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"
)

type UsersRepo struct {
	db database.DatabaseConnection
}

// NOTE: Depends on: [none]
func InitRepo(connection database.DatabaseConnection) *UsersRepo {
	logger := slog.NewLogger("users repo")
	_, err := connection.Exec(createTable)
	assert.NoError(err, logger, "Failed to create Users table")

	return &UsersRepo{connection}
}

const createTable = `CREATE TABLE IF NOT EXISTS users (
	[id] INTEGER PRIMARY KEY,
   	[email] TEXT UNIQUE,
   	[full_name] TEXT DEFAULT "",
    [session_id] TEXT DEFAULT "",
	[session_start_time] INT DEFAULT 0,
    [password_hash] TEXT DEFAULT "",
	[is_deleted] INTEGER DEFAULT 0,
	[is_admin] INTEGER DEFAULT 0,
	[create_date] INT
);`

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

const insertUser = `INSERT INTO users ([email], [full_name], [password_hash], [create_date]) VALUES (?, ?, ?, ?)`

func (r *UsersRepo) Create(user models.User) (int64, error) {
	result, err := r.db.Exec(insertUser, user.Email, user.FullName, user.PasswordHash, database.SqLiteNow().Unix())
	if err != nil {
		return 0, err
	}
	insertedId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return insertedId, nil
}

const updateUserDetails = `UPDATE users SET [full_name] = ? WHERE id = ?`

func (r *UsersRepo) UpdateDetails(user models.User) error {
	_, err := r.db.Exec(updateUserDetails, user.FullName, user.Id)
	if err != nil {
		return err
	}
	return nil
}

const updateUserPassword = `UPDATE users SET [password_hash] = ? WHERE id = ?`

func (r *UsersRepo) UpdatePassword(user models.User) error {
	_, err := r.db.Exec(updateUserPassword, user.PasswordHash, user.Id)
	if err != nil {
		return err
	}
	return nil
}

const updateUserSession = `UPDATE users SET [session_id] = ?, [session_start_time] = ? WHERE id = ?`

func (r *UsersRepo) UpdateSession(userId int64, sessionId string, sessionStartTime *database.SqLiteTime) error {
	_, err := r.db.Exec(updateUserSession, sessionId, sessionStartTime.Unix(), userId)
	if err != nil {
		return err
	}
	return nil
}

const updateUserIsAdmin = `UPDATE users SET [is_admin] = ? WHERE id = ?`

func (r *UsersRepo) UpdateIsAdmin(user models.User) error {
	_, err := r.db.Exec(updateUserIsAdmin, helpers.Btoi(user.IsAdmin), user.Id)
	if err != nil {
		return err
	}
	return nil
}

const deleteUser = `UPDATE users SET [is_deleted] = 1 WHERE id = ?`

func (r *UsersRepo) Delete(user models.User) error {
	_, err := r.db.Exec(deleteUser, user.Id)
	if err != nil {
		return err
	}
	return nil
}

const getUserByEmail = `SELECT [id], [email], [full_name], [session_id], [session_start_time], [is_admin], [is_deleted], [create_date] FROM users WHERE email = ?`

func (r *UsersRepo) GetUserByEmail(name string) (models.User, error) {
	row := r.db.QueryRow(getUserByEmail, name)

	temp := models.User{}
	err := scanUserFromRow(row, &temp)
	if err != nil {
		// TODO: A couple places rely on this error to determine if a user exitst.
		// What if the scan fails for another reason
		return models.User{}, err
	}
	return temp, nil
}

const getUserPassword = `SELECT [password_hash] FROM [users] WHERE id = ?`

func (r *UsersRepo) GetUserPassword(id int64) (string, error) {
	row := r.db.QueryRow(getUserPassword, id)
	var password string
	err := row.Scan(&password)
	if err != nil {
		// TODO: A couple places rely on this error to determine if a user exitst.
		// What if the scan fails for another reason
		return "", err
	}
	return password, nil
}

const getUser = "SELECT [id], [email], [full_name], [session_id], [session_start_time], [is_admin], [is_deleted], [create_date] FROM users WHERE id = ?"

func (r *UsersRepo) GetUserById(id int64) (models.User, error) {
	row := r.db.QueryRow(getUser, id)
	temp := models.User{}
	err := scanUserFromRow(row, &temp)
	if err != nil {
		return models.User{}, err
	}
	return temp, nil
}

const getAllUsersNotDeleted = "SELECT [id], [email], [full_name], [session_id], [session_start_time], [is_deleted],[is_admin], [create_date] FROM users WHERE is_deleted=0"

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

const logoutUser = `UPDATE users SET [session_id] = "", [session_start_time] = "" WHERE id = ?`

func (r *UsersRepo) Logout(userId int64) error {
	_, err := r.db.Exec(logoutUser, userId)
	return err
}
