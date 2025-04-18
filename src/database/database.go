package database

import (
	"database/sql"
	"fmt"
	"go-do-the-thing/src/helpers/assert"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

type DatabaseConnection struct {
	Connection *sql.DB
}

var source = "Database"

func Init(name string) DatabaseConnection {
	dbname := "./" + name + ".db"
	db, err := sql.Open("sqlite3", dbname)
	assert.NoError(err, source, "failed to start the database")

	return DatabaseConnection{db}
}

func (db DatabaseConnection) QueryRow(query string, args ...any) *sql.Row {
	assert.NotNil(db.Connection, source, "the db connection should not be nil")
	return db.Connection.QueryRow(query, args...)
}

func (db DatabaseConnection) Query(query string, args ...any) (*sql.Rows, error) {
	assert.NotNil(db.Connection, source, "the db connection should not be nil")
	return db.Connection.Query(query, args...)
}

func (db DatabaseConnection) Exec(query string, args ...any) (sql.Result, error) {
	assert.NotNil(db.Connection, source, "the db connection should not be nil")
	return db.Connection.Exec(query, args...)
}

func (db DatabaseConnection) DoesColumnExistOnTable(table, column string) (bool, error) {
	assert.NotNil(db.Connection, source, "the db connection should not be nil")
	queryString := fmt.Sprintf(checkIfColumnExists, table, column)
	query, err := db.Connection.Prepare(queryString)
	if err != nil {
		return false, err
	}
	defer query.Close()

	var countString string
	err = query.QueryRow().Scan(&countString)
	if err != nil {
		return false, err
	}
	count, err := strconv.Atoi(countString)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (db DatabaseConnection) AddColumnToTable(tableName, columnName, columnType string) error {
	assert.NotNil(db.Connection, source, "the db connection should not be nil")
	update := fmt.Sprintf(migrationAddColumn, tableName, columnName, columnType)

	exists, err := db.DoesColumnExistOnTable(tableName, columnName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = db.Connection.Exec(update)

	return err
}

const (
	migrationAddColumn  = `ALTER TABLE %s ADD %s %s`
	migrationDropColumn = `ALTER TABLE %s DROP %s`
	checkIfColumnExists = `SELECT COUNT(*) AS CNTREC FROM pragma_table_info('%s') WHERE name='%s'`
)
