package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// TestStoredProcs tests the execution of task-related stored procedures
func TestStoredProcs() error {
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
	return nil
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

	fieldDescriptions := rows.FieldDescriptions()
	colCount := len(fieldDescriptions)

	// Print header
	for i, fd := range fieldDescriptions {
		if i > 0 {
			fmt.Print(" | ")
		}
		fmt.Print(fd.Name)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", 20*colCount))

	rowCount := 0
	vals := make([]interface{}, colCount)
	scanArgs := make([]interface{}, colCount)
	for i := range vals {
		scanArgs[i] = &vals[i]
	}
	for rows.Next() {
		rowCount++
		if err := rows.Scan(scanArgs...); err != nil {
			fmt.Printf("❌ Error scanning row: %v\n", err)
			continue
		}
		for i, v := range vals {
			if i > 0 {
				fmt.Print(" | ")
			}
			fmt.Printf("%v", v)
		}
		fmt.Println()
	}
	fmt.Printf("Rows returned: %d\n", rowCount)
}

// TestFunction tests user-related functions with sample data
func TestFunction() error {
	conn := ConnectDB()
	defer conn.Close(context.Background())

	fmt.Println("Connected to database successfully")

	fmt.Println("Testing sp_get_user_by_email function...")

	var count int
	err := conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}
	fmt.Printf("Total users in database: %d\n", count)

	rows, err := conn.Query(context.Background(), "SELECT * FROM sp_get_user_by_email($1)", "m1992steen@gmail.com")
	if err != nil {
		return fmt.Errorf("failed to call sp_get_user_by_email: %w", err)
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	colCount := len(fieldDescriptions)

	// Print header
	for i, fd := range fieldDescriptions {
		if i > 0 {
			fmt.Print(" | ")
		}
		fmt.Print(fd.Name)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", 20*colCount))

	rowCount := 0
	vals := make([]interface{}, colCount)
	scanArgs := make([]interface{}, colCount)
	for i := range vals {
		scanArgs[i] = &vals[i]
	}
	for rows.Next() {
		rowCount++
		if err := rows.Scan(scanArgs...); err != nil {
			fmt.Printf("❌ Error scanning row: %v\n", err)
			continue
		}
		for i, v := range vals {
			if i > 0 {
				fmt.Print(" | ")
			}
			fmt.Printf("%v", v)
		}
		fmt.Println()
	}
	fmt.Printf("Rows returned: %d\n", rowCount)
	return nil
}

// TestAnyProc dynamically tests any stored procedure with string parameters from CLI
func TestAnyProc(procName string, params []string) error {
	conn := ConnectDB()
	defer conn.Close(context.Background())

	fmt.Printf("Testing procedure: %s with params: %v\n", procName, params)

	placeholders := make([]string, len(params))
	for i := range params {
		placeholders[i] = "$" + strconv.Itoa(i+1)
	}
	query := "SELECT * FROM " + procName
	if len(params) > 0 {
		query += "(" + strings.Join(placeholders, ", ") + ")"
	} else {
		query += "()"
	}

	args := make([]interface{}, len(params))
	for i, p := range params {
		if iVal, err := strconv.Atoi(p); err == nil {
			args[i] = iVal
		} else if fVal, err := strconv.ParseFloat(p, 64); err == nil {
			args[i] = fVal
		} else {
			args[i] = p
		}
	}

	rows, err := conn.Query(context.Background(), query, args...)
	if err != nil {
		return fmt.Errorf("❌ Error calling %s: %w", procName, err)
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	colCount := len(fieldDescriptions)

	for i, fd := range fieldDescriptions {
		if i > 0 {
			fmt.Print(" | ")
		}
		fmt.Print(fd.Name)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", 20*colCount))

	rowCount := 0
	vals := make([]interface{}, colCount)
	scanArgs := make([]interface{}, colCount)
	for i := range vals {
		scanArgs[i] = &vals[i]
	}
	for rows.Next() {
		rowCount++
		if err := rows.Scan(scanArgs...); err != nil {
			fmt.Printf("❌ Error scanning row: %v\n", err)
			continue
		}
		for i, v := range vals {
			if i > 0 {
				fmt.Print(" | ")
			}
			fmt.Printf("%v", v)
		}
		fmt.Println()
	}
	fmt.Printf("Rows returned: %d\n", rowCount)
	return nil
}

// TestAnyFunction dynamically tests any function with string parameters from CLI
func TestAnyFunction(funcName string, params []string) error {
	conn := ConnectDB()
	defer conn.Close(context.Background())

	fmt.Printf("Testing function: %s with params: %v\n", funcName, params)

	placeholders := make([]string, len(params))
	for i := range params {
		placeholders[i] = "$" + strconv.Itoa(i+1)
	}
	query := "SELECT * FROM " + funcName
	if len(params) > 0 {
		query += "(" + strings.Join(placeholders, ", ") + ")"
	} else {
		query += "()"
	}

	args := make([]interface{}, len(params))
	for i, p := range params {
		if iVal, err := strconv.Atoi(p); err == nil {
			args[i] = iVal
		} else if fVal, err := strconv.ParseFloat(p, 64); err == nil {
			args[i] = fVal
		} else {
			args[i] = p
		}
	}

	rows, err := conn.Query(context.Background(), query, args...)
	if err != nil {
		return fmt.Errorf("❌ Error calling %s: %w", funcName, err)
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	colCount := len(fieldDescriptions)

	for i, fd := range fieldDescriptions {
		if i > 0 {
			fmt.Print(" | ")
		}
		fmt.Print(fd.Name)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", 20*colCount))

	rowCount := 0
	vals := make([]interface{}, colCount)
	scanArgs := make([]interface{}, colCount)
	for i := range vals {
		scanArgs[i] = &vals[i]
	}
	for rows.Next() {
		rowCount++
		if err := rows.Scan(scanArgs...); err != nil {
			fmt.Printf("❌ Error scanning row: %v\n", err)
			continue
		}
		for i, v := range vals {
			if i > 0 {
				fmt.Print(" | ")
			}
			fmt.Printf("%v", v)
		}
		fmt.Println()
	}
	fmt.Printf("Rows returned: %d\n", rowCount)
	return nil
}
