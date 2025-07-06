package users_repo

import (
	"fmt"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"
	"time"

	"github.com/jackc/pgx/v5"
)

type UsersRepo struct {
	db     database.DatabaseConnection
	logger slog.Logger
}

var repoName = "UsersRepo"

// NOTE: Depends on: [none]
func InitRepo(connection database.DatabaseConnection) *UsersRepo {
	//TODO: Cleanup
	//_, err := connection.Exec(createTable)
	//assert.NoError(err, repoName, "Failed to create Users table")

	logger := slog.NewLogger(repoName)
	return &UsersRepo{db: connection, logger: logger}
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
		&user.IsDeleted,
		&user.IsAdmin,
		&user.IsEnabled,
		&user.CreateDate,
		&user.AccessGrantedBy,
	)
}

const insertUser = `SELECT sp_insert_user($1, $2, $3, $4)`

func (r *UsersRepo) Create(user *models.User) (int64, error) {
	r.logger.Debug("Create called - sql: %s, params: %+v", insertUser, user)
	var id int64
	err := r.db.QueryRow(insertUser, user.Email, user.FullName, user.PasswordHash, time.Now()).Scan(&id)
	if err != nil {
		r.logger.Error(err, "failed to insert user - sql: %s, params: %+v", insertUser, user)
		return 0, errors.New(errors.ErrDBInsertFailed, "failed to insert user: %w", err)
	}
	r.logger.Info("User created successfully - id: %d, email: %s", id, user.Email)
	return id, nil
}

const updateUserDetails = `SELECT sp_update_user_details($1, $2)`

func (r *UsersRepo) UpdateDetails(user models.User) error {
	r.logger.Debug("UpdateDetails called - sql: %s, params: %+v", updateUserDetails, user)
	_, err := r.db.Exec(updateUserDetails, user.FullName, user.Id)
	if err != nil {
		r.logger.Error(err, "failed to update user - sql: %s, params: %+v", updateUserDetails, user)
		return errors.New(errors.ErrDBUpdateFailed, "failed to update user: %w", err)
	}
	r.logger.Info("User details updated successfully - id: %d, fullName: %s", user.Id, user.FullName)
	return nil
}

const updateUserPassword = `SELECT sp_update_user_password($1, $2)`

func (r *UsersRepo) UpdatePassword(user models.User) error {
	r.logger.Debug("UpdatePassword called - sql: %s, params: %+v", updateUserPassword, user)
	_, err := r.db.Exec(updateUserPassword, user.PasswordHash, user.Id)
	if err != nil {
		r.logger.Error(err, "failed to set user password - sql: %s, params: %+v", updateUserPassword, user)
		return fmt.Errorf("failed to set user password: %w", err)
	}
	r.logger.Info("User password updated successfully - id: %d", user.Id)
	return nil
}

const updateUserSession = `SELECT sp_update_user_session($1, $2, $3)`

func (r *UsersRepo) UpdateSession(userId int64, sessionId string, sessionStartTime *time.Time) error {
	r.logger.Debug("UpdateSession called - sql: %s, params: %v", updateUserSession, []any{userId, sessionId, sessionStartTime})
	_, err := r.db.Exec(updateUserSession, userId, sessionId, sessionStartTime)
	if err != nil {
		r.logger.Error(err, "failed to set user session - sql: %s, params: %v", updateUserSession, []any{userId, sessionId, sessionStartTime})
		return fmt.Errorf("failed to set user session: %w", err)
	}
	r.logger.Debug("UpdateSession succeeded - id: %d", userId)
	return nil
}

const updateUserIsAdmin = `SELECT sp_update_user_is_admin($1, $2)`

func (r *UsersRepo) UpdateIsAdmin(user models.User) error {
	r.logger.Debug("UpdateIsAdmin called - sql: %s, params: %+v", updateUserIsAdmin, user)
	_, err := r.db.Exec(updateUserIsAdmin, user.IsAdmin, user.Id)
	if err != nil {
		r.logger.Error(err, "failed to set user as admin - sql: %s, params: %+v", updateUserIsAdmin, user)
		return fmt.Errorf("failed to set user as admin: %w", err)
	}
	r.logger.Info("User admin status updated successfully - id: %d, isAdmin: %t", user.Id, user.IsAdmin)
	return nil
}

const deleteUser = `SELECT sp_delete_user($1)`

func (r *UsersRepo) Delete(user models.User) error {
	r.logger.Debug("Delete called - sql: %s, params: %+v", deleteUser, user)
	_, err := r.db.Exec(deleteUser, user.Id)
	if err != nil {
		r.logger.Error(err, "failed to delete user - sql: %s, params: %+v", deleteUser, user)
		return fmt.Errorf("failed to delete user: %w", err)
	}
	r.logger.Info("User deleted successfully - id: %d", user.Id)
	return nil
}

const getUserByEmail = `SELECT * FROM sp_get_user_by_email($1)`

func (r *UsersRepo) GetUserByEmail(name string) (*models.User, error) {
	r.logger.Debug("GetUserByEmail called - sql: %s, params: %s", getUserByEmail, name)

	// Log the exact SQL being executed
	executedSQL := fmt.Sprintf("SELECT * FROM sp_get_user_by_email('%s')", name)
	r.logger.Debug("Executing SQL - sql: %s", executedSQL)

	row := r.db.QueryRow(getUserByEmail, name)

	temp := &models.User{}
	err := scanUserFromRow(row, temp)
	if err != nil {
		r.logger.Error(err, "failed to get user by email - sql: %s, params: %s, executed_sql: %s", getUserByEmail, name, executedSQL)
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	r.logger.Debug("GetUserByEmail succeeded - id: %d, params: %s", temp.Id, name)
	return temp, nil
}

const getUserPassword = `SELECT sp_get_user_password($1)`

func (r *UsersRepo) GetUserPassword(id int64) (string, error) {
	r.logger.Debug("GetUserPassword called - sql: %s, params: %d", getUserPassword, id)
	row := r.db.QueryRow(getUserPassword, id)
	var password string
	err := row.Scan(&password)
	if err != nil {
		r.logger.Error(err, "failed to get user password - sql: %s, params: %d", getUserPassword, id)
		return "", fmt.Errorf("failed to get user password: %w", err)
	}
	r.logger.Debug("GetUserPassword succeeded - id: %d", id)
	return password, nil
}

const getUser = `SELECT * FROM sp_get_user_by_id($1)`

func (r *UsersRepo) GetUserById(id int64) (*models.User, error) {
	r.logger.Debug("GetUserById called - sql: %s, params: %d", getUser, id)
	row := r.db.QueryRow(getUser, id)

	temp := &models.User{}
	err := scanUserFromRow(row, temp)
	if err != nil {
		r.logger.Error(err, "failed to get user - sql: %s, params: %d", getUser, id)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	r.logger.Debug("GetUserById succeeded - id: %d", id)
	return temp, nil
}

const getAllUsersNotDeleted = `SELECT * FROM sp_get_users_not_deleted()`

func (r *UsersRepo) GetUsers() ([]models.User, error) {
	r.logger.Debug("GetUsers called - sql: %s", getAllUsersNotDeleted)
	rows, err := r.db.Query(getAllUsersNotDeleted)
	if err != nil {
		r.logger.Error(err, "query failed to get users - sql: %s", getAllUsersNotDeleted)
		return nil, fmt.Errorf("query failed to get users: %w", err)
	}

	users := make([]models.User, 0)
	for rows.Next() {
		user := models.User{}

		err := scanUsersFromRows(rows, &user)
		if err != nil {
			r.logger.Error(err, "scan failed to get users")
			return nil, fmt.Errorf("scan failed to get users: %w", err)
		}
		users = append(users, user)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error(err, "rows.Err() in GetUsers")
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	r.logger.Debug("GetUsers succeeded - count: %d", len(users))
	return users, nil
}

const logoutUser = `SELECT sp_logout_user($1)`

func (r *UsersRepo) Logout(userId int64) error {
	r.logger.Debug("Logout called - sql: %s, params: %d", logoutUser, userId)
	_, err := r.db.Exec(logoutUser, userId)
	if err != nil {
		r.logger.Error(err, "failed to end user session - sql: %s, params: %d", logoutUser, userId)
		return fmt.Errorf("failed to end user session: %w", err)
	}
	r.logger.Info("User logged out successfully - id: %d", userId)
	return nil
}

const activateUser = `SELECT sp_update_user_is_enabled($1, $2)`

func (r *UsersRepo) ActivateUser(id int64) error {
	r.logger.Debug("ActivateUser called - sql: %s, params: %d", getUser, id)
	_, err := r.db.Exec(activateUser, id)
	if err != nil {
		r.logger.Error(err, "failed to activate user - sql: %s, params: %d", activateUser, id)
		return fmt.Errorf("failed to activate user: %w", err)
	}
	r.logger.Debug("ActivateUser succeeded - id: %d", id)
	return nil
}

const getInactiveUsers = `SELECT * FROM sp_get_users_inactive()`

func (r *UsersRepo) GetInactiveUsers() ([]models.User, error) {
	r.logger.Debug("GetInactiveUsers called - sql: %s", getInactiveUsers)
	rows, err := r.db.Query(getInactiveUsers)
	if err != nil {
		r.logger.Error(err, "query failed to get inactive users - sql: %s", getInactiveUsers)
		return nil, err
	}
	users := make([]models.User, 0)
	for rows.Next() {
		user := models.User{}
		err := scanUsersFromRows(rows, &user)
		if err != nil {
			r.logger.Error(err, "scan failed to get inactive users")
			return nil, err
		}
		users = append(users, user)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error(err, "rows.Err() in GetInactiveUsers")
		return nil, err
	}
	r.logger.Debug("GetInactiveUsers succeeded - count: %d", len(users))
	return users, nil
}
