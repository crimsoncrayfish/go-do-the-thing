package database

import (
	"context"
	"go-do-the-thing/src/helpers/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseConnection struct {
	pool *pgxpool.Pool
}

var (
	source = "Database"
	log    = logger.NewLogger(source)
)

func Init(connectionString, clientId string) *DatabaseConnection {
	conf, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		log.Fatal("Failed to parse connection string", "error", err)
	}
	if clientId != "" {
		conf.ConnConfig.User = clientId
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), conf)
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}
	if pool == nil {
		log.Fatal("Database connection is nil")
	}
	log.Info("Database setup complete")
	return &DatabaseConnection{pool}
}

func (db *DatabaseConnection) Close() {
	db.pool.Close()
}

func (db *DatabaseConnection) QueryRow(query string, args ...any) pgx.Row {
	if db.pool == nil {
		log.Fatal("Database connection is nil")
	}
	return db.pool.QueryRow(context.Background(), query, args...)
}

func (db *DatabaseConnection) Query(query string, args ...any) (pgx.Rows, error) {
	if db.pool == nil {
		log.Fatal("Database connection is nil")
	}
	return db.pool.Query(context.Background(), query, args...)
}

func (db *DatabaseConnection) Exec(query string, args ...any) (pgconn.CommandTag, error) {
	if db.pool == nil {
		log.Fatal("Database connection is nil")
	}
	return db.pool.Exec(context.Background(), query, args...)
}
