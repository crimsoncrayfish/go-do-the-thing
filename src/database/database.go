package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseConnection struct {
	pool *pgxpool.Pool
}

var source = "Database"

func Init(connectionString, clientId string) DatabaseConnection {
	conf, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		slog.Error("Failed to parse connection string", "error", err, "source", source)
		panic(fmt.Sprintf("Database initialization failed: %v", err))
	}
	if clientId != "" {
		conf.ConnConfig.User = clientId
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), conf)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err, "source", source)
		panic(fmt.Sprintf("Database connection failed: %v", err))
	}
	return DatabaseConnection{pool}
}

func (db DatabaseConnection) Close() {
	db.pool.Close()
}

func (db DatabaseConnection) QueryRow(query string, args ...any) pgx.Row {
	if db.pool == nil {
		slog.Error("Database connection is nil", "source", source)
		panic("Database connection is nil")
	}
	return db.pool.QueryRow(context.Background(), query, args...)
}

func (db DatabaseConnection) Query(query string, args ...any) (pgx.Rows, error) {
	if db.pool == nil {
		slog.Error("Database connection is nil", "source", source)
		panic("Database connection is nil")
	}
	return db.pool.Query(context.Background(), query, args...)
}

func (db DatabaseConnection) Exec(query string, args ...any) (pgconn.CommandTag, error) {
	if db.pool == nil {
		slog.Error("Database connection is nil", "source", source)
		panic("Database connection is nil")
	}
	return db.pool.Exec(context.Background(), query, args...)
}
