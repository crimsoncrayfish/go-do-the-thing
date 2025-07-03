package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBold   = "\033[1m"
)

// ApplyMigrationFile applies a single SQL migration file to the database
func ApplyMigrationFile(path string) error {
	migrationSQL, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("%s%s[ERROR]%s Failed to read migration file %s%s: %v%s\n", colorRed, colorBold, colorReset, colorBold, path, colorReset, err)
		return fmt.Errorf("failed to read migration file %s: %w", path, err)
	}

	conn := ConnectDB()
	defer conn.Close(context.Background())

	fmt.Printf("%sApplying migration:%s %s%s%s\n", colorYellow, colorReset, colorBold, path, colorReset)
	_, err = conn.Exec(context.Background(), string(migrationSQL))
	if err != nil {
		fmt.Printf("%s%s[FAILED]%s Migration %s%s%s: %v%s\n", colorRed, colorBold, colorReset, colorBold, path, colorReset, err, colorReset)
		return fmt.Errorf("failed to execute migration %s: %w", path, err)
	}

	fmt.Printf("%s%s[SUCCESS]%s Migration %s%s%s applied successfully!%s\n", colorGreen, colorBold, colorReset, colorBold, path, colorReset, colorReset)
	return nil
}

// ApplyMigrationsInDir applies all .sql migration files in a directory, in filename order
func ApplyMigrationsInDir(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("%s%s[ERROR]%s Failed to read migrations directory: %v%s\n", colorRed, colorBold, colorReset, err, colorReset)
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}

	if len(sqlFiles) == 0 {
		fmt.Printf("%sNo .sql migration files found in directory.%s\n", colorYellow, colorReset)
		return nil
	}

	sort.Strings(sqlFiles)

	successCount := 0
	total := len(sqlFiles)
	for _, fname := range sqlFiles {
		path := filepath.Join(dir, fname)
		if err := ApplyMigrationFile(path); err != nil {
			fmt.Printf("%s%s[STOPPED]%s Migration process halted due to error.%s\n", colorRed, colorBold, colorReset, colorReset)
			fmt.Printf("%sSummary: %d/%d migrations applied successfully.%s\n", colorYellow, successCount, total, colorReset)
			return err
		}
		successCount++
	}

	fmt.Printf("%s%s[COMPLETE]%s All %d migrations applied successfully!%s\n", colorGreen, colorBold, colorReset, total, colorReset)
	return nil
}
