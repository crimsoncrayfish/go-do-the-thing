package tools

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

// Database connection string
const connectionString = "postgres://admin:admin@localhost:5432/todo_db?sslmode=disable"

// ConnectDB establishes a database connection
func ConnectDB() *pgx.Conn {
	conn, err := pgx.Connect(context.Background(), connectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return conn
}
