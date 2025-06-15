package roles_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/models"

	"github.com/jackc/pgx/v5"
)

type RolesRepo struct {
	database database.DatabaseConnection
}

var repoName = "Roles Repo"

// NOTE: Depends on: []
// READONLY REPO
func InitRepo(database database.DatabaseConnection) *RolesRepo {
	//TODO: Cleanup
	//_, err := database.Exec(createRolesTable)
	//assert.NoError(err, repoName, "Failed to create Roles table")
	//_, err = database.Exec(seedRolesTable)
	//assert.NoError(err, repoName, "Failed to seed Roles table")
	return &RolesRepo{
		database: database,
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

const getAllRoles = `SELECT id, name, description FROM roles`

func (r *RolesRepo) GetAll() (roles []models.Role, err error) {
	rows, err := r.database.Query(getAllRoles)
	if err != nil {
		return nil, errors.New(errors.ErrDBReadFailed, "failed to query roles: %w", err)
	}
	defer rows.Close()

	roles = make([]models.Role, 0)
	for rows.Next() {
		role := models.Role{}

		err = scanRoleFromRows(rows, &role)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.New(errors.ErrDBGenericError, "error while iterating roles: %w", err)
	}
	return roles, nil
}
