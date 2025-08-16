package roles_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/logger"
	"go-do-the-thing/src/models"

	"github.com/jackc/pgx/v5"
)

type RolesRepo struct {
	database database.DatabaseConnection
	logger   *logger.Logger
}

// NOTE: Depends on: []
// READONLY REPO
func InitRepo(database database.DatabaseConnection) *RolesRepo {
	log := logger.NewLogger("RolesRepo")
	return &RolesRepo{
		database: database,
		logger:   log,
	}
}

func scanRoleFromRows(rows pgx.Rows, item *models.Role) error {
	err := rows.Scan(
		&item.Id,
		&item.Name,
		&item.Description,
	)
	if err != nil {
		return errors.NewDbGenericError("failed to scan role from rows", err, "", nil)
	}
	return nil
}

const getAllRoles = `SELECT * FROM sp_get_all_roles()`

func (r *RolesRepo) GetAll() (roles []models.Role, err error) {
	rows, err := r.database.Query(getAllRoles)
	if err != nil {
		r.logger.Error("failed to query roles", "error", err, "sql", getAllRoles)
		return nil, errors.NewDbReadFailed("roles", err, "", nil)
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
		return nil, errors.NewDbGenericError("rows error in GetAll", err, "", nil)
	}
	r.logger.Debug("GetAll succeeded", "count", len(roles))
	return roles, nil
}
