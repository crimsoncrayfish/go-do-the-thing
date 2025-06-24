# Dev Tools

This directory contains development utilities for managing the todo application database and stored procedures.

## Prerequisites

- Go 1.23.0+
- PostgreSQL database running on localhost:5432
- Database credentials: admin/admin
- Database name: todo_db
- `pg_dump` command available (for backup tool)

## Quick Start

### Using Runner Scripts (Recommended)

**Windows PowerShell:**
```powershell
.\run_tool.ps1 list-procs
.\run_tool.ps1 backup
.\run_tool.ps1 health-check
```

**Windows Command Prompt:**
```cmd
run_tool.bat list-procs
run_tool.bat backup
run_tool.bat health-check
```

### Direct Execution

```bash
# Navigate to dev_tools directory
cd dev_tools

# List all stored procedures
go run main.go list-procs

# View table schema
go run main.go print-schema

# Test stored procedures
go run main.go test-procs

# Show help
go run main.go
```

## Available Tools

### 🔍 **Database Schema Tools**

- **`list-procs`** - Lists all stored procedures with their signatures and return types
- **`print-schema`** - Displays the schema of all tables in the public schema

### 🧪 **Testing Tools**

- **`test-procs`** - Tests the execution of task-related stored procedures (currently static, can be made dynamic)
- **`test-function`** - Tests user-related functions with sample data (currently static, can be made dynamic)

### 🛡️ **Safety & Validation Tools**

- **`backup`** - Creates a timestamped database backup using pg_dump
- **`health-check`** - Validates data integrity, checks for orphaned records, and reports inconsistencies

## Tool Categories

### Safe Tools (No Data Loss Risk)
- `list-procs` - Read-only schema inspection
- `print-schema` - Read-only table structure
- `test-procs` - Read-only procedure testing
- `test-function` - Read-only function testing
- `health-check` - Read-only data validation
- `backup` - Creates backups (safe)

## Safety Notes

⚠️ **Warning**: Always backup your database before making schema or data changes.

### Recommended Workflow

1. **Before making changes:**
   ```bash
   go run main.go backup
   ```

2. **Check current state:**
   ```bash
   go run main.go list-procs
   go run main.go health-check
   ```

3. **Make changes:**
   - Use SQL migration files and apply them using the general-purpose migration functions in Go (see below).

4. **Verify changes:**
   ```bash
   go run main.go test-procs
   go run main.go health-check
   ```

## Running Migrations (General-Purpose Tool)

Migrations are now managed via Go functions, not via CLI. To apply migrations:

```
// In Go code:
ApplyMigrationFile("../src/database/migrations/20240614-update-id-columns.sql")
ApplyMigrationsInDir("../src/database/migrations")
```

Add these calls to a custom tool or run them from a Go script as needed.

## Environment Setup

Make sure your PostgreSQL database is running and accessible with the credentials specified in each tool. You can modify the connection string in each tool if your setup differs.

### Database Connection Details
- Host: localhost
- Port: 5432
- User: admin
- Password: admin
- Database: todo_db

## Troubleshooting

### Common Issues

1. **Connection refused**: Ensure PostgreSQL is running on localhost:5432
2. **Authentication failed**: Verify username/password are correct
3. **pg_dump not found**: Install PostgreSQL client tools
4. **Permission denied**: Ensure you have access to the database

### Getting Help

If a tool fails, check:
1. Database connection (try `list-procs` first)
2. Database permissions
3. Tool-specific error messages
4. Backup before retrying destructive operations

## Tool Descriptions

### Schema Tools
- **list-procs**: Shows all stored procedures with their parameters and return types
- **print-schema**: Displays table structure for all tables in the public schema

### Testing Tools  
- **test-procs**: Executes task-related stored procedures with sample data to verify they work
- **test-function**: Tests user lookup functions with a specific email address

### Safety Tools
- **backup**: Creates a timestamped SQL backup of the entire database
- **health-check**: Validates data integrity and reports any issues found

## File Structure

```
dev_tools/
├── main.go                  # Entry point for all tools
├── tools/
│   ├── database.go          # Database connection utilities
│   ├── schema.go            # Schema inspection tools
│   ├── testing.go           # Testing tools
│   ├── migrations.go        # Migration helpers (for use in Go code)
│   └── safety.go            # Backup and validation tools
├── go.mod                   # Go module file
├── run_tool.ps1             # PowerShell runner script
├── run_tool.bat             # Batch runner script
└── README.md                # This documentation
```

## Architecture

The dev tools are organized in a modular structure:

- **`tools/database.go`**: Shared database connection utilities
- **`tools/schema.go`**: Schema inspection and listing tools
- **`tools/testing.go`**: Testing and validation tools
- **`tools/migrations.go`**: Database migration helpers
- **`tools/safety.go`**: Backup and data integrity tools

This modular design makes the codebase:
- **Maintainable**: Each tool category is in its own file
- **Extensible**: Easy to add new tools to appropriate categories
- **Readable**: Clear separation of concerns
- **Testable**: Individual functions can be tested in isolation

## Running Migrations

You can apply migrations using the general-purpose migration functions in Go:

```
// Apply a single migration file
ApplyMigrationFile("../src/database/migrations/20240614-update-id-columns.sql")

// Apply all migrations in a directory (in filename order)
ApplyMigrationsInDir("../src/database/migrations")
```

Add these calls to a custom tool or run them from a Go script as needed 