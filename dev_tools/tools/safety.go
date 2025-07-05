package tools

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// CreateBackup creates a timestamped database backup using pg_dump
func CreateBackup() error {
	// Parse connection string from database.go
	u, err := url.Parse(connectionString)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	user := u.User.Username()
	password, _ := u.User.Password()
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "5432"
	}
	dbName := strings.TrimPrefix(u.Path, "/")

	// Create backup filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupFilename := fmt.Sprintf("backup_%s_%s.sql", dbName, timestamp)

	fmt.Printf("Creating database backup: %s\n", backupFilename)

	// Construct pg_dump command
	cmd := exec.Command("pg_dump",
		"-h", host,
		"-p", port,
		"-U", user,
		"-d", dbName,
		"--clean",
		"--if-exists",
		"--create",
		"--no-password",
		"-f", backupFilename,
	)

	// Set environment variable for password
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))

	// Execute the backup
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create backup: %w\nOutput: %s", err, output)
	}

	fmt.Printf("✅ Database backup created successfully: %s\n", backupFilename)

	// Verify the backup file exists and has content
	if fileInfo, err := os.Stat(backupFilename); err == nil {
		fmt.Printf("Backup file size: %d bytes\n", fileInfo.Size())
	} else {
		fmt.Printf("Warning: Could not verify backup file: %v\n", err)
	}
	return nil
}

// DataHealthCheck validates data integrity, checks for orphaned records, and reports inconsistencies
func DataHealthCheck() error {
	conn := ConnectDB()
	defer conn.Close(context.Background())

	fmt.Println("🔍 Database Health Check")
	fmt.Println("========================")

	// Check table counts
	checkTableCounts(conn)

	// Check for orphaned records
	checkOrphanedRecords(conn)

	// Check data consistency
	checkDataConsistency(conn)

	fmt.Println("\n✅ Health check completed!")
	return nil
}

// checkTableCounts checks the record count for all tables
func checkTableCounts(conn *pgx.Conn) {
	fmt.Println("\n📊 Table Record Counts:")

	// Dynamically get all table names in the public schema
	tableRows, err := conn.Query(context.Background(), `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name;
	`)
	if err != nil {
		fmt.Printf("  ❌ Error querying table names: %v\n", err)
		return
	}
	defer tableRows.Close()

	tables := []string{}
	for tableRows.Next() {
		var tableName string
		err := tableRows.Scan(&tableName)
		if err != nil {
			fmt.Printf("  ❌ Error scanning table name: %v\n", err)
			continue
		}
		tables = append(tables, tableName)
	}

	if len(tables) == 0 {
		fmt.Println("  No tables found in public schema.")
		return
	}

	for _, table := range tables {
		var count int
		err := conn.QueryRow(context.Background(), "SELECT COUNT(*) FROM "+table).Scan(&count)
		if err != nil {
			fmt.Printf("  %s: ❌ Error - %v\n", table, err)
		} else {
			fmt.Printf("  %s: %d records\n", table, count)
		}
	}
}

// checkOrphanedRecords checks for orphaned records in the database
func checkOrphanedRecords(conn *pgx.Conn) {
	fmt.Println("\n🔗 Checking for orphaned records:")

	// Items without valid projects
	var orphanedItems int
	err := conn.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM items i 
		LEFT JOIN projects p ON i.project_id = p.id 
		WHERE i.project_id IS NOT NULL AND p.id IS NULL
	`).Scan(&orphanedItems)
	if err != nil {
		fmt.Printf("  ❌ Error checking orphaned items: %v\n", err)
	} else if orphanedItems > 0 {
		fmt.Printf("  ⚠️  Found %d items with invalid project_id\n", orphanedItems)
	} else {
		fmt.Printf("  ✅ No orphaned items found\n")
	}

	// Project users without valid users
	var orphanedProjectUsers int
	err = conn.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM project_users pu 
		LEFT JOIN users u ON pu.user_id = u.id 
		WHERE u.id IS NULL
	`).Scan(&orphanedProjectUsers)
	if err != nil {
		fmt.Printf("  ❌ Error checking orphaned project_users: %v\n", err)
	} else if orphanedProjectUsers > 0 {
		fmt.Printf("  ⚠️  Found %d project_users with invalid user_id\n", orphanedProjectUsers)
	} else {
		fmt.Printf("  ✅ No orphaned project_users found\n")
	}
}

// checkDataConsistency checks for data consistency issues
func checkDataConsistency(conn *pgx.Conn) {
	fmt.Println("\n🔧 Checking data consistency:")

	// Items with future complete dates
	var futureCompleteItems int
	err := conn.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM items 
		WHERE complete_date > NOW() AND complete_date IS NOT NULL
	`).Scan(&futureCompleteItems)
	if err != nil {
		fmt.Printf("  ❌ Error checking future complete dates: %v\n", err)
	} else if futureCompleteItems > 0 {
		fmt.Printf("  ⚠️  Found %d items with future complete_date\n", futureCompleteItems)
	} else {
		fmt.Printf("  ✅ No items with future complete dates\n")
	}

	// Items with due_date before created_date
	var inconsistentDates int
	err = conn.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM items 
		WHERE due_date < created_date AND due_date IS NOT NULL
	`).Scan(&inconsistentDates)
	if err != nil {
		fmt.Printf("  ❌ Error checking date consistency: %v\n", err)
	} else if inconsistentDates > 0 {
		fmt.Printf("  ⚠️  Found %d items with due_date before created_date\n", inconsistentDates)
	} else {
		fmt.Printf("  ✅ No date inconsistencies found\n")
	}
}
