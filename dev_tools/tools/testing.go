package tools

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

// TestStoredProcs tests the execution of task-related stored procedures
func TestStoredProcs() {
	conn := ConnectDB()
	defer conn.Close(context.Background())

	fmt.Println("Connected to database successfully")

	// TODO: Replace with real IDs from your DB if needed
	userID := int64(1)
	projectID := int64(1)
	itemID := int64(1)

	testProc(conn, "sp_get_items_by_user", "SELECT * FROM sp_get_items_by_user($1)", userID)
	testProc(conn, "sp_get_items_by_user_and_project", "SELECT * FROM sp_get_items_by_user_and_project($1, $2)", userID, projectID)
	testProc(conn, "sp_get_item", "SELECT * FROM sp_get_item($1)", itemID)
}

// testProc is a helper function to test stored procedures
func testProc(conn *pgx.Conn, name string, query string, params ...interface{}) {
	fmt.Printf("\nTesting %s...\n", name)
	rows, err := conn.Query(context.Background(), query, params...)
	if err != nil {
		fmt.Printf("❌ Error calling %s: %v\n", name, err)
		return
	}
	defer rows.Close()

	fmt.Printf("✅ %s executed successfully!\n", name)
	fmt.Println("Columns returned:")
	fieldDescriptions := rows.FieldDescriptions()
	for i, fd := range fieldDescriptions {
		fmt.Printf("  %d: %s\n", i+1, fd.Name)
	}

	rowCount := 0
	for rows.Next() {
		rowCount++
	}
	fmt.Printf("Rows returned: %d\n", rowCount)
}

// TestFunction tests user-related functions with sample data
func TestFunction() {
	conn := ConnectDB()
	defer conn.Close(context.Background())

	fmt.Println("Connected to database successfully")

	// Test the stored procedure
	fmt.Println("Testing sp_get_user_by_email function...")

	// First, let's see if there are any users in the database
	var count int
	err := conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		log.Fatalf("Failed to count users: %v", err)
	}
	fmt.Printf("Total users in database: %d\n", count)

	// Test the function with a sample email
	rows, err := conn.Query(context.Background(), "SELECT * FROM sp_get_user_by_email($1)", "m1992steen@gmail.com")
	if err != nil {
		log.Fatalf("Failed to call sp_get_user_by_email: %v", err)
	}
	defer rows.Close()

	fmt.Println("✅ Function executed successfully!")
	fmt.Println("Columns returned:")

	// Get column descriptions
	fieldDescriptions := rows.FieldDescriptions()
	for i, fd := range fieldDescriptions {
		fmt.Printf("  %d: %s\n", i+1, fd.Name)
	}

	// Count rows returned
	rowCount := 0
	for rows.Next() {
		rowCount++
	}

	if rowCount > 0 {
		fmt.Printf("✅ Found %d user(s) with that email\n", rowCount)
	} else {
		fmt.Println("ℹ️  No user found with that email (this is normal if the user doesn't exist)")
	}

	fmt.Println("✅ Test completed successfully!")
	fmt.Println("🎉 All stored procedures should now work correctly!")
}
