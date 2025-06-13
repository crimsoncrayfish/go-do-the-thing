package roles_repo

import (
	"database/sql"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/models"
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

func scanRoleFromRows(rows *sql.Rows, item *models.Role) error {
	return rows.Scan(
		&item.Id,
		&item.Name,
		&item.Description,
	)
}

const getAllRoles = `SELECT id, name, description FROM roles`

func (r *RolesRepo) GetAll() (roles []models.Role, err error) {
	rows, err := r.database.Query(getAllRoles)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err = rows.Close()
	}(rows)

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
		return nil, err
	}
	return roles, nil
}
