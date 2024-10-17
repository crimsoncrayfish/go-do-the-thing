package roles_repo

import (
	"database/sql"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"
)

type RolesRepo struct {
	database database.DatabaseConnection
}

// NOTE: Depends on: []
// READONLY REPO
func InitRepo(database database.DatabaseConnection) *RolesRepo {
	logger := slog.NewLogger("roles repo")
	_, err := database.Exec(createRolesTable)
	assert.NoError(err, logger, "Failed to create Roles table")
	_, err = database.Exec(seedRolesTable)
	assert.NoError(err, logger, "Failed to seed Roles table")
	return &RolesRepo{
		database: database,
	}
}

const createRolesTable = `CREATE TABLE IF NOT EXISTS roles (
	[id] INTEGER PRIMARY KEY,
   	[name] TEXT DEFAULT '' NOT NULL,
   	[Description] TEXT DEFAULT '' NOT NULL
);`
const seedRolesTable = `INSERT OR IGNORE INTO roles (id, name, description) VALUES
	(1, 'Big boss', 'Project Administrator.'),
	(2, 'Little boss', 'Can create, assign and complete tasks as well as add/remove users from the project.'),
	(3, 'Grunt', 'Can create, assign and complete tasks.'),
	(4, 'Pleb', 'Can complete tasks.')
	`

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
