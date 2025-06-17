package database

import (
	"context"
	"go-do-the-thing/src/helpers/assert"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseConnection struct {
	pool *pgxpool.Pool
}

var source = "Database"

func Init(connectionString string) DatabaseConnection {
	conf, err := pgxpool.ParseConfig(connectionString)
	assert.NoError(err, source, "Failed to parse connectionstring")
	pool, err := pgxpool.NewWithConfig(context.Background(), conf)
	assert.NoError(err, source, "Failed to connect to db")
	return DatabaseConnection{pool}
}

func (db DatabaseConnection) Close() {
	db.pool.Close()
}

func (db DatabaseConnection) QueryRow(query string, args ...any) pgx.Row {
	assert.NotNil(db.pool, source, "the db connection should not be nil")
	return db.pool.QueryRow(context.Background(), query, args...)
}

func (db DatabaseConnection) Query(query string, args ...any) (pgx.Rows, error) {
	assert.NotNil(db.pool, source, "the db connection should not be nil")
	return db.pool.Query(context.Background(), query, args...)
}

func (db DatabaseConnection) Exec(query string, args ...any) (pgconn.CommandTag, error) {
	assert.NotNil(db.pool, source, "the db connection should not be nil")
	return db.pool.Exec(context.Background(), query, args...)
}
