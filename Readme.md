# Go-Do-The-Thing

A modern todo application built with Go, HTMX, and Tailwind CSS, featuring user authentication, project management, and task organization.

## Features

- **User Authentication**: JWT-based authentication with session management
- **Project Management**: Organize tasks into projects with user permissions
- **Task Management**: Create, edit, and track tasks with due dates and priorities
- **Admin Controls**: Administrative interface for user management
- **Responsive Design**: Mobile-friendly interface built with Tailwind CSS
- **Real-time Updates**: HTMX-powered dynamic UI without page refreshes
- **Minimal Dependencies**: Lightweight frontend with no large JavaScript frameworks

## Architecture

This application follows a clean layered architecture:
- **Handlers**: HTTP request handling and routing
- **Services**: Business logic and validation
- **Repositories**: Data access layer with PostgreSQL
- **Models**: Domain models and form validation
- **Middleware**: Authentication, logging, and rate limiting

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

## Prerequisites

### Required Tools
- **Go 1.21+**: Main programming language
- **Node.js & npm**: For CSS building and package management
- **Docker & Docker Compose**: For local database setup
- **PostgreSQL**: Database (via Docker or local installation)

### Development Tools
```bash
# Hot reloading during development
go install github.com/air-verse/air@latest

# Template generation
go install github.com/a-h/templ/cmd/templ@latest

# Database migrations
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Frontend Dependencies
```bash
# Install npm dependencies for Tailwind CSS
npm install

# Note: HTMX is currently loaded via CDN to minimize build complexity
# Small custom JavaScript functions are in static/js/ for specific needs only
```

## Design Philosophy

This project prioritizes **minimal dependencies** and **simplicity**:

- **HTMX-First**: Dynamic behavior through HTMX attributes rather than JavaScript frameworks
- **Vanilla JavaScript**: Small, focused functions in `static/js/` only when HTMX cannot handle the requirement
- **No Large Libraries**: Avoids jQuery, React, Vue, or other heavy client-side frameworks
- **Progressive Enhancement**: HTML-first approach with minimal JavaScript enhancement
- **Server-Side Rendering**: Leverages Go templates (Templ) for UI logic

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

## Quick Start

### 1. Clone and Setup
```bash
git clone <repository-url>
cd go-do-the-thing
npm install
```

### 2. Start Local Database
```bash
docker compose -f docker-compose-db-local.yml up -d
```

### 3. Run Migrations
```bash
docker run -v ./src/database/migrations:/migrations --network go-do-the-thing_default migrate/migrate -path=/migrations/ -database postgres://admin:admin@db:5432/todo_db?sslmode=disable up
```

### 4. Generate Templates and Build CSS
```bash
templ generate
npm run build
```

### 5. Start Development Server
```bash
air  # Hot reloading
# OR
go run main.go  # Standard run
```

The application will be available at `https://localhost:8080` (development uses HTTPS).

## Environment Variables

### Development
- Uses hardcoded database connection for local Docker setup
- HTTPS certificates: `server.crt` and `server.key` required in root directory

### Production
- `DATABASE_URL`: PostgreSQL connection string
- `PORT`: Server port (default: 8080)
- `ENV`: Set to "production"
- `AZURE_CLIENT_ID`: For Azure authentication (optional)

## Project Status

### Completed Features
- [x] Home page and navigation
- [x] User authentication (JWT-based)
- [x] User registration and login
- [x] Task management (CRUD operations)
- [x] Project organization
- [x] Admin user controls
- [x] Responsive design
- [x] Docker containerization
- [x] Database migrations
- [x] Templ template system

### Planned Features
- [ ] Tags on projects and tasks
- [ ] Time logging on tasks
- [ ] CSV export functionality
- [ ] Dashboard with statistics
- [ ] Enhanced search and filtering

## Contributing

1. Follow the patterns established in `AGENTS.md`
2. Run tests before submitting changes
3. Ensure migrations have both UP and DOWN variants
4. Update documentation for new features

## Troubleshooting

### Windows Development
If compilation issues occur on Windows:
```cmd
$env:GOTMPDIR = "PATH TO TEMP DIR"
go env -w CGO_ENABLED=1
go env -w CC="zig cc"
```
Ensure ZIG is installed.
