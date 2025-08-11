# Go Do The Thing - Lightweight Task Management

A minimal, dependency-light task management application built with Go, HTMX, Templ, and Tailwind CSS. This project prioritizes simplicity, performance, and minimal JavaScript usage while providing a modern, responsive user experience.

## Project Philosophy

This application is designed with the following principles:
- **Minimal External Dependencies**: Keep third-party dependencies to an absolute minimum
- **Server-Side Rendering**: Leverage Go's Templ for HTML generation and HTMX for dynamic interactions
- **Minimal JavaScript**: Only 3 custom JS components (colorpicker, datepicker, dialog) for essential UI functionality
- **Lightweight Architecture**: Fast, efficient, and easy to deploy

## Features

- **User Management**: Registration, login, JWT authentication, admin controls
- **Task Management**: Create, edit, delete, and organize tasks
- **Project Organization**: Group tasks into projects with tagging support
- **Responsive Design**: Mobile-first design with Tailwind CSS
- **Real-time Updates**: HTMX-powered dynamic UI updates without page refreshes
- **Database Migrations**: Structured PostgreSQL schema with migration support
- **Containerized Deployment**: Docker support for easy deployment

## APP Demo
LOGIN:

https://github.com/user-attachments/assets/a649ccc3-6481-4f9c-9339-7399afbbc51b

TASKS:

https://github.com/user-attachments/assets/83962fef-6d57-456a-8b15-bb3b1741efaf

https://github.com/user-attachments/assets/eabc63f6-8925-40fe-a873-24f382ead6eb

RESPONSIVE DESIGN:

https://github.com/user-attachments/assets/019db6ad-4cd1-4796-8665-d72cd45a4e12

https://github.com/user-attachments/assets/6a16eeb7-5d24-4a81-9319-715e0ea25b3c

## Migrations

All migrations are in `.sql` files in the `/src/database/migrations/` folder

Each migration should have an UP and a DOWN variant. To apply migrations use the [Golang Migrate Tool](github.com/golang-migrate/migrate/v4)

Usage:
```cmd
--Get the tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
--Create new migrations file
migrate create -ext sql -dir db/migrations -seq create_users_table
--Apply migrations
migrate -database ${POSTGRESQL_URL} -path db/migrations up
```

For local development the `docker-compose-db-local.yml` file was created. (If on windows run the following commands in a wsl instance)

```
--Create the db and apply migrations
docker compose -f docker-compose-db-local.yml up -d
--Tear down and remove the db
docker compose -f docker-compose-db-local.yml down -v

--Run up migration
docker run -v ./src/database/migrations:/migrations --network go-do-the-thing_default migrate/migrate -path=/migrations/ -database postgres://admin:admin@db:5432/todo_db?sslmode=disable up
--Run down migration
docker run -v ./src/database/migrations:/migrations --network go-do-the-thing_default migrate/migrate -path=/migrations/ -database postgres://admin:admin@db:5432/todo_db?sslmode=disable down -all
```

## Containerized

There is a DOCKERFILE in the root directory for building this project to run in the cloud.

Use the following commands to test the dockerfile locally
```
--Build docker container
docker build -t todo-app:local .
--Run (remove -rm if it crashes and you need logs)
docker run -rm -d -p 8080:8080 --network go-do-the-thing_default --name todo-app-container --env-file .env todo-app:local
--Extra
docker stop todo-app-container
docker logs -f todo-app-container
docker rm todo-app-container
```

## Project Structure

```
├── src/
│   ├── database/           # Database layer and migrations
│   │   ├── migrations/     # SQL migration files
│   │   └── repos/         # Repository pattern implementations
│   ├── handlers/          # HTTP request handlers
│   │   ├── admin/         # Admin functionality
│   │   ├── home/          # Home page
│   │   ├── project/       # Project management
│   │   ├── task/          # Task management
│   │   └── users/         # User authentication
│   ├── helpers/           # Utility functions and constants
│   ├── middleware/        # HTTP middleware (auth, logging, rate limiting)
│   ├── models/            # Data models and form definitions
│   ├── services/          # Business logic layer
│   └── shared/            # Shared Templ templates
├── static/
│   ├── css/               # Tailwind CSS files
│   ├── js/                # Minimal custom JavaScript components
│   └── img/               # Static assets
└── dev_tools/             # Development utilities and database tools
```

## Technology Stack

- **Backend**: Go with standard library HTTP server
- **Frontend**: HTMX + Templ templates + Tailwind CSS
- **Database**: PostgreSQL with migration support
- **Authentication**: Custom JWT implementation
- **Deployment**: Docker containerization

## Environment Configuration

### Required Environment Variables

The application requires the following environment variables to be configured:

#### Production Environment Variables
For production deployments, create a `.env` file in the root directory with:

```bash
# Application Environment
ENV=production

# Server Configuration
PORT=8080

# Database Configuration
DATABASE_URL=postgres://username:password@host:port/database_name?sslmode=require

# JWT Authentication Keys (Base64 encoded)
JWT_PRIVATE_KEY=<base64-encoded-private-key>
JWT_PUBLIC_KEY=<base64-encoded-public-key>

# Optional: Azure Integration
AZURE_CLIENT_ID=<your-azure-client-id>
```

#### Development Environment Variables
For local development, create a `.env` file with:

```bash
# Application Environment (optional - defaults to development)
ENV=development

# Server Configuration (optional - defaults to 8080)
PORT=8080

# Database Configuration (required for local development)
DATABASE_URL=postgres://admin:admin@localhost:5432/todo_db?sslmode=disable

# JWT Authentication Keys (Base64 encoded) - Required
JWT_PRIVATE_KEY=<base64-encoded-private-key>
JWT_PUBLIC_KEY=<base64-encoded-public-key>

# Optional: Azure Integration
AZURE_CLIENT_ID=<your-azure-client-id>
```

You'll also need a `.dev.env` file for migration tasks:
```bash
# Development Environment for Migrations
DATABASE_URL=postgres://admin:admin@localhost:5432/todo_db?sslmode=disable
```

### Generating JWT Keys

To generate the required JWT keys for authentication:

```bash
# Generate private key
openssl genpkey -algorithm RSA -out private.key -outform PEM -pkeyopt rsa_keygen_bits:2048

# Generate public key from private key
openssl rsa -pubout -in private.key -out public.key

# Convert to base64 for environment variables
JWT_PRIVATE_KEY=$(base64 -i private.key)
JWT_PUBLIC_KEY=$(base64 -i public.key)
```

### Environment Files to Create

1. **`.env`** (root directory) - Main environment configuration file for the application
2. **`.dev.env`** (root directory) - Environment file specifically for local migration tasks
3. **`keys/private.key`** (optional) - JWT private key file (also stored as base64 in .env)
4. **`keys/public.key`** (optional) - JWT public key file (also stored as base64 in .env)

**Note**: The `task setup:env` command creates both `.env` and `.dev.env` files automatically. Alternatively, you can copy the provided `.env.example` file and customize it manually.

## Getting Started

### Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd go-do-the-thing
   ```

2. **Complete project initialization**
   
   **Option A: Using Taskfile (Recommended)**
   ```bash
   # Install Taskfile first if not already installed
   go install github.com/go-task/task/v3/cmd/task@latest
   
   # Complete setup: dependencies + environment with JWT keys
   task setup:init
   ```
   
   **Option B: Step-by-step setup**
   ```bash
   # Install Taskfile first if not already installed
   go install github.com/go-task/task/v3/cmd/task@latest
   
   # Install dependencies only
   task install:deps
   
   # Create .env file with generated JWT keys
   task setup:env
   ```

3. **Manual installation (if not using Taskfile)**
   ```bash
   # Configure environment variables manually (using provided example)
   cp .env.example .env
   
   # Generate JWT keys and update .env file
   openssl genpkey -algorithm RSA -out keys/private.key -outform PEM -pkeyopt rsa_keygen_bits:2048
   openssl rsa -pubout -in keys/private.key -out keys/public.key
   
   # Convert to base64 and add to .env file
   JWT_PRIVATE_KEY=$(base64 -i keys/private.key | tr -d '\n')
   JWT_PUBLIC_KEY=$(base64 -i keys/public.key | tr -d '\n')
   
   # Install Go dependencies
   go mod download
   
   # Install development tools
   go install github.com/a-h/templ/cmd/templ@latest
   go install github.com/air-verse/air@latest
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   
   # Install Node.js dependencies for Tailwind CSS
   npm install
   ```

3. **Start the PostgreSQL database**
   ```bash
   # Start local PostgreSQL database using Docker
   docker compose -f docker-compose-db-local.yml up -d
   
   # Verify database is running
   docker compose -f docker-compose-db-local.yml ps
   ```

4. **Run database migrations**
   
   **Option A: Using Taskfile (Recommended)**
   ```bash
   # Install Taskfile if not already installed
   go install github.com/go-task/task/v3/cmd/task@latest
   
   # Run migrations
   task local:migrate -- up
   
   # Or migrate to specific version
   task local:migrate:goto V=3
   ```
   
   **Option B: Manual (No Taskfile dependency)**
   ```bash
   # Install migrate tool
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   
   # Run migrations manually
   docker run -v ./src/database/migrations:/migrations --network go-do-the-thing_default migrate/migrate -path=/migrations/ -database postgres://admin:admin@db:5432/todo_db?sslmode=disable up
   ```

5. **Start the application**
   ```bash
   # Build CSS and start development server with hot reload
   npm run build
   air
   ```

6. **Access the application**
   - Open your browser to `http://localhost:8080`
   - Register a new account or use existing credentials

### Custom JavaScript Components

The application uses only 3 minimal JavaScript components:
- **ColorPicker**: Custom color selection for projects/tags
- **DatePicker**: Date selection for task due dates
- **Dialog**: Modal dialogs for confirmations and forms

## Development Setup

### Prerequisites

Before starting development, ensure you have the following installed:

- **Go** (1.23.0 or later)
- **Node.js** and **npm** (for Tailwind CSS)
- **Docker** and **Docker Compose** (for PostgreSQL database)

**Quick Setup with Taskfile:**
```bash
# Install Taskfile
go install github.com/go-task/task/v3/cmd/task@latest

# Complete project initialization (dependencies + environment)
task setup:init

# Or run individual tasks:
# task install:deps    # Install dependencies only
# task setup:env       # Create .env with JWT keys only
```

**Manual Setup (if you prefer not to use Taskfile):**
```bash
# Required Go tools
go install github.com/a-h/templ/cmd/templ@latest
go install github.com/air-verse/air@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Node.js dependencies (for Tailwind CSS and HTMX)
npm install
```

### Running the Application

```bash
# Start the PostgreSQL database (required)
docker compose -f docker-compose-db-local.yml up -d

# Run migrations
docker run -v ./src/database/migrations:/migrations --network go-do-the-thing_default migrate/migrate -path=/migrations/ -database postgres://admin:admin@db:5432/todo_db?sslmode=disable up

# Start development server with hot reload
air

# Build CSS (in separate terminal)
npm run watch
```

## Compilation issues on a Windows machine

If there are compilation issues one of these is likely to fix it
```cmd
$env:GOTMPDIR = "PATH TO TEMP DIR"
go env -w CGO_ENABLED=1
go env -w CC="zig cc"
```
Ensure ZIG is installed on the pc

## Generate Public/Private key example
```cmd
ssh-keygen -t rsa -b 4096

--For JWTs (i.e. in keys/)
openssl genpkey -algorithm RSA -out private.key -outform PEM -pkeyopt rsa_keygen_bits:2048
openssl rsa -pubout -in private.key -out public.key

--FOR HTTPS
openssl genpkey -algorithm RSA -out private_key.pem -outform PEM -pkeyopt rsa_keygen_bits:2048
openssl req -new -x509 -key private_key.pem -out certificate.pem -days 365
```

## HTMX is wierd
- Cant process <body></body> as an oob swap
- oob swaps need to be before the main swap if there are 2 things being swopped
- Cant get nested oob swaps to work :(
- oob swaps with rows are [wierd](https://htmx.org/attributes/hx-swap-oob/)

## Development Roadmap

### Completed Features ✅
- [x] Home page with responsive design
- [x] Complete task management (create, edit, delete, list)
- [x] Custom JWT authentication system
- [x] User registration and login
- [x] User-task associations
- [x] Templ template integration
- [x] Secure logout functionality
- [x] Task editing security improvements
- [x] Admin controls for user management
- [x] Docker containerization
- [x] Project-based task organization
- [x] Database migration system
- [x] HTMX integration for dynamic UI
- [x] Minimal JavaScript component system

### Planned Features 🚧
- [ ] Enhanced tagging system for projects and tasks
- [ ] Time tracking and logging on tasks
- [ ] CSV export functionality
- [ ] Analytics dashboard with task statistics
- [ ] Task search and filtering
- [ ] Task priority levels
- [ ] Due date notifications
- [ ] Project collaboration features

### Technical Improvements 🔧
- [ ] API documentation
- [ ] Comprehensive test coverage
- [ ] Performance optimization
- [ ] Enhanced error handling
- [ ] Accessibility improvements
