package main

import (
	"fmt"
	"os"

	"github.com/m1992/go-do-the-thing/dev_tools/tools"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	tool := os.Args[1]

	switch tool {
	case "list-procs":
		if err := tools.ListAllProcs(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "print-schema":
		if err := tools.PrintTableColumns(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "run-migrations":
		if err := tools.ApplyMigrationsInDir("../src/database/migrations"); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "backup":
		if err := tools.CreateBackup(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "health-check":
		if err := tools.DataHealthCheck(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("❌ Unknown tool: %s\n", tool)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Dev Tools - Database Management Utilities")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Usage: go run main.go <tool-name> [args]")
	fmt.Println()
	fmt.Println("Available tools:")
	fmt.Println("  list-procs           - List all stored procedures")
	fmt.Println("  print-schema         - Print table schema")
	fmt.Println("  test-procs           - Test stored procedures (static)")
	fmt.Println("  test-function        - Test user functions (static)")
	fmt.Println("  test-proc <proc> [params...] - Dynamically test any stored procedure")
	fmt.Println("  test-func <func> [params...] - Dynamically test any function")
	fmt.Println("  backup               - Create database backup")
	fmt.Println("  health-check         - Check data health")
	fmt.Println("  run-migrations       - Apply all SQL migrations in src/database/migrations")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run main.go list-procs")
	fmt.Println("  go run main.go print-schema")
	fmt.Println("  go run main.go test-proc sp_get_items_by_user 1")
	fmt.Println("  go run main.go test-func sp_get_user_by_email user@example.com")
	fmt.Println("  go run main.go backup")
	fmt.Println("  go run main.go health-check")
}
