package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func Init(name string) (*sql.DB, error) {
	dbname := "./" + name + ".db"
	db, err := sql.Open("sqlite3", dbname)
	if err != nil {
		return nil, err
	}
	return db, nil
}
