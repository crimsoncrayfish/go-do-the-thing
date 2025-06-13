package database

import (
	"database/sql"
	"go-do-the-thing/src/helpers/assert"

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
