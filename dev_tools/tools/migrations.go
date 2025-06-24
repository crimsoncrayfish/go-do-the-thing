package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ApplyMigrationFile applies a single SQL migration file to the database
func ApplyMigrationFile(path string) error {
	migrationSQL, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %w", path, err)
	}

	conn := ConnectDB()
	defer conn.Close(context.Background())

	fmt.Printf("Applying migration: %s\n", path)
	_, err = conn.Exec(context.Background(), string(migrationSQL))
	if err != nil {
		return fmt.Errorf("failed to execute migration %s: %w", path, err)
	}

	fmt.Printf("Migration %s applied successfully!\n", path)
	return nil
}

// ApplyMigrationsInDir applies all .sql migration files in a directory, in filename order
func ApplyMigrationsInDir(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}

	if len(sqlFiles) == 0 {
		fmt.Println("No .sql migration files found in directory.")
		return nil
	}

	sort.Strings(sqlFiles)

	for _, fname := range sqlFiles {
		path := filepath.Join(dir, fname)
		if err := ApplyMigrationFile(path); err != nil {
			return err
		}
	}

	fmt.Println("All migrations applied successfully!")
	return nil
}
