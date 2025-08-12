package roles_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/models"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

type RolesRepo struct {
	database database.DatabaseConnection
	logger   *slog.Logger
}

// NOTE: Depends on: []
// READONLY REPO
func InitRepo(database database.DatabaseConnection) *RolesRepo {
	logger := slog.With("source", "RolesRepo")
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
	r.logger.Debug("GetAll called", "sql", getAllRoles)
	rows, err := r.database.Query(getAllRoles)
	if err != nil {
		r.logger.Error("failed to query roles", "error", err, "sql", getAllRoles)
		return nil, errors.New(errors.ErrDBReadFailed, "failed to query roles: %w", err)
	}
	defer rows.Close()

	roles = make([]models.Role, 0)
	for rows.Next() {
		role := models.Role{}

		err = scanRoleFromRows(rows, &role)
		if err != nil {
			r.logger.Error("failed to scan row in GetAll", "error", err)
			return nil, err
		}
		roles = append(roles, role)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error("rows.Err() in GetAll", "error", err)
		return nil, errors.New(errors.ErrDBGenericError, "some error contained in the rows: %w", err)
	}
	r.logger.Debug("GetAll succeeded", "count", len(roles))
	return roles, nil
}
