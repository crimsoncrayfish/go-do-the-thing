package users_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/logger"
	"go-do-the-thing/src/models"
	"time"

	"github.com/jackc/pgx/v5"
)

type UsersRepo struct {
	db     database.DatabaseConnection
	logger *logger.Logger
}

// NOTE: Depends on: [none]
func InitRepo(connection database.DatabaseConnection) *UsersRepo {
	log := logger.NewLogger("UsersRepo")
	return &UsersRepo{
		db:     connection,
		logger: log,
	}
}

func scanUserFromRow(row pgx.Row, user *models.User) error {
	return row.Scan(
		&user.Id,
		&user.Email,
		&user.FullName,
		&user.SessionId,
		&user.SessionStartTime,
		&user.IsAdmin,
		&user.IsEnabled,
		&user.IsDeleted,
		&user.CreateDate,
		&user.AccessGrantedBy,
	)
}

func scanUsersFromRows(rows pgx.Rows, user *models.User) error {
	return rows.Scan(
		&user.Id,
		&user.Email,
		&user.FullName,
		&user.SessionId,
		&user.SessionStartTime,
		&user.IsAdmin,
		&user.IsEnabled,
		&user.IsDeleted,
		&user.CreateDate,
		&user.AccessGrantedBy,
	)
}

const insertUser = `SELECT sp_insert_user($1, $2, $3, $4)`

func (r *UsersRepo) Create(user *models.User) (int64, error) {
	var id int64
	err := r.db.QueryRow(insertUser, user.Email, user.FullName, user.PasswordHash, time.Now()).Scan(&id)
	if err != nil {
		r.logger.Error("failed to insert user", "error", err, "sql", insertUser, "email", user.Email)
		return 0, errors.NewDbInsertFailed("user", err, "", nil)
	}
	return id, nil
}

const updateUserDetails = `SELECT sp_update_user_details($1, $2)`

func (r *UsersRepo) UpdateDetails(user models.User) error {
	_, err := r.db.Exec(updateUserDetails, user.FullName, user.Id)
	if err != nil {
		r.logger.Error("failed to update user", "error", err, "sql", updateUserDetails, "user_id", user.Id, "full_name", user.FullName)
		return errors.NewDbUpdateFailed("user", err, "details", nil)
	}
	return nil
}

const updateUserPassword = `SELECT sp_update_user_password($1, $2)`

func (r *UsersRepo) UpdatePassword(user models.User) error {
	_, err := r.db.Exec(updateUserPassword, user.PasswordHash, user.Id)
	if err != nil {
		r.logger.Error("failed to set user password", "error", err, "sql", updateUserPassword, "user_id", user.Id)
		return errors.NewDbUpdateFailed("user", err, "password", nil)
	}
	return nil
}

const updateUserSession = `SELECT sp_update_user_session($1, $2, $3)`

func (r *UsersRepo) UpdateSession(userId int64, sessionId string, sessionStartTime *time.Time) error {
	_, err := r.db.Exec(updateUserSession, userId, sessionId, sessionStartTime)
	if err != nil {
		r.logger.Error("failed to set user session", "error", err, "sql", updateUserSession, "user_id", userId, "session_id", sessionId, "sessionStartTime", sessionStartTime)
		return errors.NewDbUpdateFailed("user", err, "session", nil)
	}
	return nil
}

const updateUserIsAdmin = `SELECT sp_update_user_is_admin($1, $2)`

func (r *UsersRepo) UpdateIsAdmin(user models.User) error {
	_, err := r.db.Exec(updateUserIsAdmin, user.IsAdmin, user.Id)
	if err != nil {
		r.logger.Error("failed to set user as admin", "error", err, "sql", updateUserIsAdmin, "user_id", user.Id)
		return errors.NewDbUpdateFailed("user", err, "admin status", nil)
	}
	return nil
}

const deleteUser = `SELECT sp_delete_user($1)`

func (r *UsersRepo) Delete(user models.User) error {
	_, err := r.db.Exec(deleteUser, user.Id)
	if err != nil {
		r.logger.Error("failed to delete user", "error", err, "sql", deleteUser, "user_id", user.Id)
		return errors.NewDbDeleteFailed("user", err, "", nil)
	}
	return nil
}

const getUserByEmail = `SELECT * FROM sp_get_user_by_email($1)`

func (r *UsersRepo) GetUserByEmail(email string) (*models.User, error) {

	row := r.db.QueryRow(getUserByEmail, email)

	temp := &models.User{}
	err := scanUserFromRow(row, temp)
	if err != nil {
		r.logger.Error("failed to get user by email", "error", err, "sql", getUserByEmail, "email", email)
		return nil, errors.NewDbReadFailed("user", err, "by email", nil)
	}
	return temp, nil
}

const getUserPassword = `SELECT sp_get_user_password($1)`

func (r *UsersRepo) GetUserPassword(id int64) (string, error) {
	row := r.db.QueryRow(getUserPassword, id)
	var password string
	err := row.Scan(&password)
	if err != nil {
		r.logger.Error("failed to get user password", "error", err, "sql", getUserPassword, "id", id)
		return "", errors.NewDbReadFailed("user", err, "password", nil)
	}
	return password, nil
}

const getUser = `SELECT * FROM sp_get_user_by_id($1)`

func (r *UsersRepo) GetUserById(id int64) (*models.User, error) {
	row := r.db.QueryRow(getUser, id)

	temp := &models.User{}
	err := scanUserFromRow(row, temp)
	if err != nil {
		r.logger.Error("failed to get user", "error", err, "sql", getUser, "id", id)
		return nil, errors.NewDbReadFailed("user", err, "by id", nil)
	}
	return temp, nil
}

const getAllUsersNotDeleted = `SELECT * FROM sp_get_users_not_deleted()`

func (r *UsersRepo) GetUsers() ([]models.User, error) {
	rows, err := r.db.Query(getAllUsersNotDeleted)
	if err != nil {
		r.logger.Error("query failed to get users", "error", err, "sql", getAllUsersNotDeleted)
		return nil, errors.NewDbReadFailed("users", err, "", nil)
	}

	users := make([]models.User, 0)
	for rows.Next() {
		user := models.User{}

		err := scanUsersFromRows(rows, &user)
		if err != nil {
			r.logger.Error("scan failed to get users", "error", err)
			return nil, errors.NewDbGenericError("users", err, "scan failed", nil)
		}
		users = append(users, user)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error("rows.Err() in GetUsers", "error", err)
		return nil, errors.NewDbGenericError("users", err, "rows error", nil)
	}
	return users, nil
}

const logoutUser = `SELECT sp_logout_user($1)`

func (r *UsersRepo) Logout(userId int64) error {
	_, err := r.db.Exec(logoutUser, userId)
	if err != nil {
		r.logger.Error("failed to end user session", "error", err, "sql", logoutUser, "user_id", userId)
		return errors.NewDbUpdateFailed("user", err, "session logout", nil)
	}
	return nil
}

const updateUserEnabled = `SELECT * FROM sp_update_user_is_enabled($1, $2)`

func (r *UsersRepo) UpdateUserEnabled(id int64, enabled bool) (string, error) {
	row := r.db.QueryRow(updateUserEnabled, id, enabled)
	var email string
	err := row.Scan(&email)
	if err != nil {
		r.logger.Error("failed to update user enabled status", "error", err, "sql", updateUserEnabled, "id", id, "enabled", enabled)
		return "", errors.NewDbUpdateFailed("user", err, "enabled status", nil)
	}
	return email, nil
}
