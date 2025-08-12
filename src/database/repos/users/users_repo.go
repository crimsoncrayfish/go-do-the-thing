package users_repo

import (
	"fmt"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/models"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

type UsersRepo struct {
	db     database.DatabaseConnection
	logger *slog.Logger
}

// NOTE: Depends on: [none]
func InitRepo(connection database.DatabaseConnection) *UsersRepo {
	logger := slog.With("source", "UsersRepo")
	return &UsersRepo{
		db:     connection,
		logger: logger,
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
	r.logger.Debug("Create called", "sql", insertUser, "params", user)
	var id int64
	err := r.db.QueryRow(insertUser, user.Email, user.FullName, user.PasswordHash, time.Now()).Scan(&id)
	if err != nil {
		r.logger.Error("failed to insert user", "error", err, "sql", insertUser, "email", user.Email)
		return 0, errors.New(errors.ErrDBInsertFailed, "failed to insert user: %w", err)
	}
	return id, nil
}

const updateUserDetails = `SELECT sp_update_user_details($1, $2)`

func (r *UsersRepo) UpdateDetails(user models.User) error {
	r.logger.Debug("UpdateDetails called", "sql", updateUserDetails, "params", user)
	_, err := r.db.Exec(updateUserDetails, user.FullName, user.Id)
	if err != nil {
		r.logger.Error("failed to update user", "error", err, "sql", updateUserDetails, "user_id", user.Id, "full_name", user.FullName)
		return errors.New(errors.ErrDBUpdateFailed, "failed to update user: %w", err)
	}
	return nil
}

const updateUserPassword = `SELECT sp_update_user_password($1, $2)`

func (r *UsersRepo) UpdatePassword(user models.User) error {
	r.logger.Debug("UpdatePassword called", "sql", updateUserPassword, "user_id", user.Id)
	_, err := r.db.Exec(updateUserPassword, user.PasswordHash, user.Id)
	if err != nil {
		r.logger.Error("failed to set user password", "error", err, "sql", updateUserPassword, "user_id", user.Id)
		return fmt.Errorf("failed to set user password: %w", err)
	}
	return nil
}

const updateUserSession = `SELECT sp_update_user_session($1, $2, $3)`

func (r *UsersRepo) UpdateSession(userId int64, sessionId string, sessionStartTime *time.Time) error {
	r.logger.Debug("UpdateSession called", "sql", updateUserSession, "userId", userId, "sessionId", sessionId, "sessionStartTime", sessionStartTime)
	_, err := r.db.Exec(updateUserSession, userId, sessionId, sessionStartTime)
	if err != nil {
		r.logger.Error("failed to set user session", "error", err, "sql", updateUserSession, "userId", userId, "sessionId", sessionId, "sessionStartTime", sessionStartTime)
		return fmt.Errorf("failed to set user session: %w", err)
	}
	return nil
}

const updateUserIsAdmin = `SELECT sp_update_user_is_admin($1, $2)`

func (r *UsersRepo) UpdateIsAdmin(user models.User) error {
	r.logger.Debug("UpdateIsAdmin called", "sql", updateUserIsAdmin, "user_id", user.Id)
	_, err := r.db.Exec(updateUserIsAdmin, user.IsAdmin, user.Id)
	if err != nil {
		r.logger.Error("failed to set user as admin", "error", err, "sql", updateUserIsAdmin, "user_id", user.Id)
		return fmt.Errorf("failed to set user as admin: %w", err)
	}
	return nil
}

const deleteUser = `SELECT sp_delete_user($1)`

func (r *UsersRepo) Delete(user models.User) error {
	r.logger.Debug("Delete called", "sql", deleteUser, "user_id", user.Id)
	_, err := r.db.Exec(deleteUser, user.Id)
	if err != nil {
		r.logger.Error("failed to delete user", "error", err, "sql", deleteUser, "user_id", user.Id)
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

const getUserByEmail = `SELECT * FROM sp_get_user_by_email($1)`

func (r *UsersRepo) GetUserByEmail(email string) (*models.User, error) {
	r.logger.Debug("GetUserByEmail called", "sql", getUserByEmail, "email", email)

	row := r.db.QueryRow(getUserByEmail, email)

	temp := &models.User{}
	err := scanUserFromRow(row, temp)
	if err != nil {
		r.logger.Error("failed to get user by email", "error", err, "sql", getUserByEmail, "email", email)
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return temp, nil
}

const getUserPassword = `SELECT sp_get_user_password($1)`

func (r *UsersRepo) GetUserPassword(id int64) (string, error) {
	r.logger.Debug("GetUserPassword called", "sql", getUserPassword, "id", id)
	row := r.db.QueryRow(getUserPassword, id)
	var password string
	err := row.Scan(&password)
	if err != nil {
		r.logger.Error("failed to get user password", "error", err, "sql", getUserPassword, "id", id)
		return "", fmt.Errorf("failed to get user password: %w", err)
	}
	return password, nil
}

const getUser = `SELECT * FROM sp_get_user_by_id($1)`

func (r *UsersRepo) GetUserById(id int64) (*models.User, error) {
	r.logger.Debug("GetUserById called", "sql", getUser, "id", id)
	row := r.db.QueryRow(getUser, id)

	temp := &models.User{}
	err := scanUserFromRow(row, temp)
	if err != nil {
		r.logger.Error("failed to get user", "error", err, "sql", getUser, "id", id)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return temp, nil
}

const getAllUsersNotDeleted = `SELECT * FROM sp_get_users_not_deleted()`

func (r *UsersRepo) GetUsers() ([]models.User, error) {
	r.logger.Debug("GetUsers called", "sql", getAllUsersNotDeleted)
	rows, err := r.db.Query(getAllUsersNotDeleted)
	if err != nil {
		r.logger.Error("query failed to get users", "error", err, "sql", getAllUsersNotDeleted)
		return nil, fmt.Errorf("query failed to get users: %w", err)
	}

	users := make([]models.User, 0)
	for rows.Next() {
		user := models.User{}

		err := scanUsersFromRows(rows, &user)
		if err != nil {
			r.logger.Error("scan failed to get users", "error", err)
			return nil, fmt.Errorf("scan failed to get users: %w", err)
		}
		users = append(users, user)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error("rows.Err() in GetUsers", "error", err)
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	return users, nil
}

const logoutUser = `SELECT sp_logout_user($1)`

func (r *UsersRepo) Logout(userId int64) error {
	r.logger.Debug("Logout called", "sql", logoutUser, "userId", userId)
	_, err := r.db.Exec(logoutUser, userId)
	if err != nil {
		r.logger.Error("failed to end user session", "error", err, "sql", logoutUser, "userId", userId)
		return fmt.Errorf("failed to end user session: %w", err)
	}
	return nil
}

const updateUserEnabled = `SELECT * FROM sp_update_user_is_enabled($1, $2)`

func (r *UsersRepo) UpdateUserEnabled(id int64, enabled bool) (string, error) {
	r.logger.Debug("UpdateUserEnabled called", "sql", updateUserEnabled, "id", id, "enabled", enabled)
	row := r.db.QueryRow(updateUserEnabled, id, enabled)
	var email string
	err := row.Scan(&email)
	if err != nil {
		r.logger.Error("failed to update user enabled status", "error", err, "sql", updateUserEnabled, "id", id, "enabled", enabled)
		return "", fmt.Errorf("failed to update user enabled status: %w", err)
	}
	return email, nil
}
