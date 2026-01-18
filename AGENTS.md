# Project Context for AI Agents

This document provides context for AI agents working on the "Go Do The Thing" project.

---

## Project Overview

**Name:** Go Do The Thing  
**Type:** Task management web application  
**Architecture:** Server-side rendered monolith with HTMX for interactivity

---

## Tech Stack

| Layer | Technology | Notes |
|-------|------------|-------|
| Language | Go 1.23.0+ | Standard library `net/http` for server |
| Templates | Templ | Type-safe Go HTML templates (`.templ` files generate `_templ.go`) |
| Frontend | HTMX + Tailwind CSS | Minimal custom JavaScript |
| Database | PostgreSQL | Connection pooling via `pgx/pgxpool` |
| Auth | JWT | Tokens stored in HTTP-only cookies |
| Dev Server | Air | Hot reload configured in `.air.toml` |
| Task Runner | Taskfile | Commands in `Taskfile.yml` |

### Driving Principles

- Solve everything you can with Tailwind styling and HTMX
- Isolate large chunks of JavaScript to components (like the datepicker)
- No custom CSS
- Reuse JS from the core `functions.js` file if needed
- **YAGNI (You Aren't Gonna Need It):** Only build what is immediately needed. Do not add speculative stored procedures, API endpoints, or features "just in case." Add them when there's a concrete use case.

---

## Directory Structure

```
/
├── main.go                     # Application entry point
├── go.mod / go.sum             # Go dependencies
├── package.json                # Node dependencies (Tailwind, HTMX)
├── Taskfile.yml                # Task runner commands
├── .air.toml                   # Hot reload config
├── .env.example                # Environment template
│
├── src/
│   ├── database/
│   │   ├── database.go         # PostgreSQL connection setup
│   │   ├── migrations/         # SQL migration files
│   │   └── repos/              # Repository pattern implementations
│   │       ├── repos.go        # RepoContainer (DI container)
│   │       ├── users/
│   │       ├── tasks/
│   │       ├── projects/
│   │       ├── roles/
│   │       ├── project-users/
│   │       ├── project-tags/
│   │       └── task-tags/
│   │
│   ├── handlers/               # HTTP request handlers
│   │   ├── helpers.go          # Shared handler utilities
│   │   ├── admin/
│   │   ├── home/
│   │   ├── project/
│   │   ├── task/
│   │   └── users/              # Auth (login/register/logout)
│   │
│   ├── services/               # Business logic layer
│   │   ├── admin/service.go
│   │   ├── project/service.go
│   │   ├── task/service.go
│   │   └── user/service.go
│   │
│   ├── middleware/
│   │   ├── middleware.go       # Middleware stack composition
│   │   ├── auth.go             # JWT authentication
│   │   ├── admin.go            # Admin role verification
│   │   ├── logging.go          # Request logging
│   │   └── ratelimiting.go     # IP-based rate limiting
│   │
│   ├── models/                 # Data models
│   │   ├── user.go
│   │   ├── task.go
│   │   ├── project.go
│   │   ├── tag.go
│   │   ├── roles.go
│   │   └── forms/              # Form validation models
│   │
│   ├── helpers/
│   │   ├── errors/             # Custom error handling with toast notifications
│   │   ├── security/           # JWT and password hashing (bcrypt)
│   │   ├── logger/             # Structured logging with slog
│   │   └── constants/          # Application constants
│   │
│   └── shared/templ/           # Shared Templ components
│
├── static/
│   ├── css/                    # Tailwind input/output
│   ├── js/                     # HTMX + custom JS (datepicker, colorpicker, dialog)
│   └── img/
│
└── dev_tools/                  # Development utilities
    └── tools/                  # DB, migrations, testing helpers
```

---

## Architecture Patterns

### Request Flow
```
HTTP Request → Middleware Stack → Handler → Service → Repository → Database
                                    ↓
                              Templ Template → HTTP Response
```

### Repository Pattern
- All database access goes through repositories in `src/database/repos/`
- `RepoContainer` in `repos.go` provides dependency injection
- Each repo handles one entity type (users, tasks, projects, etc.)

### Service Layer
- Business logic lives in `src/services/`
- Services receive repositories via constructor injection
- Handlers call services, services call repositories

### Middleware Stack
Composed in `main.go`:
1. Logging (all requests)
2. Rate limiting (IP-based, configurable)
3. Authentication (JWT validation)
4. Admin verification (role check for admin routes)

---

## Key Conventions

### Error Handling
- Custom error types in `src/helpers/errors/`
- Errors render as toast notifications via Templ components
- Use `errors.NewAppError()` for user-facing errors

### Authentication
- JWT tokens in HTTP-only cookies
- Password hashing with bcrypt
- Session management in users table
- Roles: Big Boss, Little Boss, Grunt, Pleb

### HTMX Patterns
- `hx-get`, `hx-post`, `hx-delete` for AJAX requests
- `hx-swap="outerHTML"` is the default swap strategy
- OOB (Out-of-Band) swaps for updating multiple elements
- Handlers return partial HTML, not full pages

### Templ Templates
- `.templ` files compile to `_templ.go`
- Run `templ generate` after modifying templates
- Shared components in `src/shared/templ/`
- Each handler package has its own templates

### Database Migrations
- Located in `src/database/migrations/`
- Run with `task migrate:up` or `task migrate:down`
- Sequential numbering: `001_`, `002_`, etc.

---

## Common Commands

```bash
# Development
task dev              # Start dev server with hot reload
task build            # Build production binary

# Database (Docker)
task db:build         # Build local Postgres image (run once)
task db:up            # Start local Postgres container
task db:down          # Stop container (keeps data)
task db:wipe          # Stop container and delete data
task db:reset         # Wipe, recreate, and run all migrations
task db:shell         # Open psql shell to local database
task db:logs          # Tail Postgres logs

# Migrations
task local:migrate -- up      # Run pending migrations (local)
task local:migrate -- down    # Rollback last migration (local)
task prod:migrate -- up       # Run migrations on production (with confirmation)

# Templates
templ generate        # Regenerate Templ files

# CSS
npx tailwindcss -i static/css/input.css -o static/css/output.css --watch
```

### First-Time Setup

```bash
# 1. Copy environment file
cp .env.example .env

# 2. Build and start the database
task db:build
task db:up

# 3. Run migrations
task local:migrate -- up

# 4. Start development server
task dev
```

---

## Environment Variables

Required in `.env` (see `.env.example`):
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - Secret for signing JWT tokens
- `PORT` - HTTP server port (default: 8080)
- `RATE_LIMIT` - Requests per minute per IP (default: 500)

---

## Database Schema

**Core Tables:**
- `users` - User accounts with session data
- `roles` - Permission roles (4 predefined)
- `projects` - Project containers
- `tasks` - Task items with status, due dates, soft delete
- `project_users` - Project membership (many-to-many)
- `project_tags` / `task_tags` - Tagging systems

---

## Frontend Notes

### Minimal JavaScript
Only 4 custom JS files:
- `functions.js` - HTMX config, toast handling
- `datepicker.js` - Date selection component
- `colorpicker.js` - Color selection component  
- `dialog.js` - Modal dialog component

### Styling
- Tailwind CSS with custom config
- Dark mode support via class strategy
- Mobile-first responsive design

---

## Testing

- Run Go tests: `go test ./...`
- Dev tools in `dev_tools/` for manual testing
- No frontend testing framework currently

---

## Adding New Features

### New Entity Checklist
1. Create model in `src/models/`
2. Add migration in `src/database/migrations/`
3. Create repository in `src/database/repos/`
4. Add to `RepoContainer` in `repos.go`
5. Create service in `src/services/`
6. Create handler in `src/handlers/`
7. Add Templ templates in handler package
8. Register routes in `main.go`

### New Handler Checklist
1. Create package in `src/handlers/`
2. Add `handler.go` with constructor and routes
3. Add `.templ` files for views
4. Run `templ generate`
5. Register in `main.go`

---

## Roadmap / Goals

### 1. Time Tracking
**Status:** Partially implemented (start button exists)

**Data Model:** Event-based time tracking
- `task_time_events` table with event types: `start`, `stop`
- Each event records task_id, user_id, timestamp, event_type
- Duration calculated from start/stop event pairs
- Allows for multiple work sessions per task

**Features:**
- "Start" button begins tracking (creates start event)
- "Stop" button pauses tracking without completing (creates stop event)
- Support for multiple start/stop cycles per task
- Calculate total time from event pairs

### 2. Admin Oversight Dashboard
**Status:** Not started

- Usage analytics for administrators
- Number of active users
- Activity metrics and reporting

### 3. Timeline Modification Component
**Status:** Partially built (component exists but not fully functional)

- Responsive drag-to-resize interface for time entries
- Visually adjust time spent on a task for a given day
- Modifies the stop event timestamp (or creates adjustment events)
- **Note:** May require app restructuring to fully support
