package roles_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"

	"github.com/jackc/pgx/v5"
)

type RolesRepo struct {
	database database.DatabaseConnection
	logger   slog.Logger
}

var repoName = "RolesRepo"

// NOTE: Depends on: []
// READONLY REPO
func InitRepo(database database.DatabaseConnection) *RolesRepo {
	logger := slog.NewLogger(repoName)
	//TODO: Cleanup
	//_, err := database.Exec(createRolesTable)
	//assert.NoError(err, repoName, "Failed to create Roles table")
	//_, err = database.Exec(seedRolesTable)
	//assert.NoError(err, repoName, "Failed to seed Roles table")
	return &RolesRepo{
		database: database,
		logger:   logger,
	}
}

func scanRoleFromRows(rows pgx.Rows, item *models.Role) error {
	err := rows.Scan(
		&item.Id,
		&item.Name,
		&item.Description,
	)
	if err != nil {
		return errors.New(errors.ErrDBGenericError, "failed to scan role from rows: %w", err)
	}
	return nil
}

const getAllRoles = `SELECT * FROM sp_get_all_roles()`

func (r *RolesRepo) GetAll() (roles []models.Role, err error) {
	r.logger.Debug("GetAll called - sql: %s", getAllRoles)
	rows, err := r.database.Query(getAllRoles)
	if err != nil {
		r.logger.Error(err, "failed to query roles - sql: %s", getAllRoles)
		return nil, errors.New(errors.ErrDBReadFailed, "failed to query roles: %w", err)
	}
	defer rows.Close()

	roles = make([]models.Role, 0)
	for rows.Next() {
		role := models.Role{}

		err = scanRoleFromRows(rows, &role)
		if err != nil {
			r.logger.Error(err, "failed to scan row in GetAll")
			return nil, err
		}
		roles = append(roles, role)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error(err, "rows.Err() in GetAll")
		return nil, errors.New(errors.ErrDBGenericError, "some error contained in the rows: %w", err)
	}
	r.logger.Debug("GetAll succeeded - count: %d", len(roles))
	return roles, nil
}
