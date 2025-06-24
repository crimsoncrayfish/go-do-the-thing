package tools

import (
	"context"
	"fmt"
)

// ListAllProcs lists all stored procedures in the database
func ListAllProcs() error {
	conn := ConnectDB()
	defer conn.Close(context.Background())

	fmt.Println("All stored procedures in the database:")
	fmt.Println("=====================================")

	query := `
	SELECT 
		proname as function_name,
		pg_get_function_arguments(oid) as arguments,
		pg_get_function_result(oid) as return_type
	FROM pg_proc 
	WHERE pronamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')
	ORDER BY proname;
	`

	rows, err := conn.Query(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to query functions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var funcName, args, returnType string
		err := rows.Scan(&funcName, &args, &returnType)
		if err != nil {
			fmt.Printf("Failed to scan row: %v\n", err)
			continue
		}
		fmt.Printf("\nFunction: %s\n", funcName)
		fmt.Printf("  Arguments: %s\n", args)
		fmt.Printf("  Returns: %s\n", returnType)
	}

	fmt.Println("\n=====================================")
	fmt.Println("Total functions listed above")
	return nil
}

// PrintTableColumns displays the schema of all tables in the public schema
func PrintTableColumns() error {
	conn := ConnectDB()
	defer conn.Close(context.Background())

	// Get all table names in the public schema
	tableRows, err := conn.Query(context.Background(), `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name;
	`)
	if err != nil {
		return fmt.Errorf("failed to query table names: %w", err)
	}
	defer tableRows.Close()

	tables := []string{}
	for tableRows.Next() {
		var tableName string
		err := tableRows.Scan(&tableName)
		if err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if len(tables) == 0 {
		fmt.Println("No tables found in public schema.")
		return nil
	}

	for _, table := range tables {
		fmt.Printf("\nTable: %s\n", table)
		fmt.Println("Column\t\tType")
		fmt.Println("------\t\t----")
		colRows, err := conn.Query(context.Background(), `
			SELECT column_name, data_type
			FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1
			ORDER BY ordinal_position;
		`, table)
		if err != nil {
			return fmt.Errorf("failed to query columns for table %s: %w", table, err)
		}
		for colRows.Next() {
			var column, dtype string
			err := colRows.Scan(&column, &dtype)
			if err != nil {
				return fmt.Errorf("failed to scan column: %w", err)
			}
			fmt.Printf("%s\t%s\n", column, dtype)
		}
		colRows.Close()
	}
	return nil
}
