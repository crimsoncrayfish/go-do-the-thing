# AI Assistant Instructions for Go-Do-The-Thing Project

This document provides comprehensive guidelines for AI assistants working on this Go-based todo application built with HTMX, Tailwind CSS, and PostgreSQL.

## Project Architecture Overview

This project follows a layered architecture pattern with a **minimal dependency philosophy**:
- **Handlers**: HTTP request handling and routing
- **Services**: Business logic layer
- **Repositories**: Data access layer
- **Models**: Data structures and forms
- **Middleware**: Cross-cutting concerns (auth, logging, rate limiting)

### Core Philosophy: Minimal Dependencies
- **HTMX-First**: Use HTMX for dynamic behavior instead of heavy JavaScript frameworks
- **Vanilla JavaScript**: Small, focused functions in `static/js/` for specific needs only
- **No Large Libraries**: Avoid jQuery, React, Vue, or other large client-side frameworks
- **Progressive Enhancement**: HTML-first approach with minimal JavaScript enhancement
- **Server-Side Rendering**: Leverage Go templates (Templ) for most UI logic

## Database Layer Standards

### Migrations
- All migrations are stored in `src/database/migrations/`
- Each migration MUST have both UP and DOWN variants
- Naming convention: `XXXXX_description.up.sql` and `XXXXX_description.down.sql`
- Use sequential numbering (00001, 00002, etc.)
- Always test both UP and DOWN migrations
- Use the golang-migrate tool for migration management

### Repository Pattern
- Each entity has its own repository in `src/database/repos/{entity}/`
- Repository files are named `{entity}_repo.go`
- All repositories are initialized through `repos.go` container
- Repository methods should be focused on data access only
- Use prepared statements and proper error handling
- Repository methods should return domain models, not database-specific types

### Repository Container
- All repositories are managed through `RepoContainer` in `src/database/repos/repos.go`
- Use getter methods to access repositories (e.g., `GetUsersRepo()`)
- Initialize repositories in `NewContainer()` function
- Follow the established pattern for adding new repositories

## Service Layer Standards

### Service Structure
- Services are located in `src/services/{domain}/`
- Each service file is named `service.go`
- Services contain business logic and orchestrate repository calls
- Services should validate input and handle business rules
- Use dependency injection through the setup function pattern: `Setup{Domain}Service(repo_container *repos.RepoContainer)`

### Service Patterns
- Always include structured logging using the custom slog package
- Set a `serviceSource` constant for logging context
- Handle errors appropriately and log them with context
- Services should return domain models or view models
- Validate business rules before calling repositories
- Use transactions when multiple repository operations are needed

### Error Handling
- Use custom error types from `src/helpers/errors/`
- Log errors with appropriate context (user ID, email, etc.)
- Don't expose internal errors to the client
- Return generic error messages for security-sensitive operations

## Handler Layer Standards

### Handler Structure
- Handlers are organized by domain in `src/handlers/{domain}/`
- Main handler logic in `handler.go`
- Templ templates in `templ/` subdirectory
- Use the setup function pattern: `Setup{Domain}Handler()`

### HTTP Handling Patterns
- Use `AcceptHeaderSwitch()` helper for JSON/HTML content negotiation
- HTMX-specific redirects: use `Redirect()` for HX-Location, `RedirectFull()` for HX-Redirect
- Always validate user permissions before processing requests
- Extract user context from middleware-injected values
- Use proper HTTP status codes

### Form Handling
- Forms are defined in `src/models/forms/`
- All forms must implement the error interface pattern
- Forms should have `GetErrors()` and `SetError()` methods
- Use form validation before calling services
- Return form with errors for client-side display

## Model Standards

### Domain Models
- Core models in `src/models/`
- Separate view models for client presentation
- Use proper Go naming conventions
- Include JSON tags for API responses
- Models should be immutable where possible

### Form Models
- Located in `src/models/forms/`
- Must implement error handling interface
- Include both data and validation errors
- Use constructor functions (e.g., `NewTaskForm()`)
- Provide default values where appropriate

## Middleware Standards

### Middleware Pattern
- All middleware in `src/middleware/`
- Use the standard middleware signature: `func(http.Handler) http.Handler`
- Middleware stacking through `CreateStack()` function
- Order matters: rate limiting → logging → authentication → authorization

### Available Middleware
- **Authentication**: JWT-based user authentication
- **Admin**: Admin role verification
- **Logging**: Request/response logging
- **Rate Limiting**: Request rate limiting

### Middleware Usage
- `middleware_admin`: Full stack with admin verification
- `middleware_full`: Full stack with authentication
- `middleware_no_auth`: Optional authentication for public endpoints

## Security Standards

### Authentication
- JWT-based authentication using RSA keys
- Session management with database storage
- Password hashing using bcrypt
- Secure session invalidation on logout

### Authorization
- Role-based access control (admin/user)
- Project-level permissions for tasks
- Always verify user ownership of resources
- Use middleware for consistent authorization

### Security Best Practices
- Never log passwords or sensitive data
- Use prepared statements to prevent SQL injection
- Validate all user input
- Implement proper CORS policies
- Use HTTPS in production

## HTMX Integration Standards

### HTMX-First Philosophy
- **Minimal Dependencies**: Avoid large JavaScript libraries; rely primarily on HTMX for dynamic behavior
- **Custom JavaScript**: Small, focused functions in `static/js/` for specific needs
- **Progressive Enhancement**: HTML-first approach with HTMX enhancing functionality
- **No Framework Lock-in**: Keep JavaScript minimal and vanilla where possible

### HTMX Patterns
- Use `hx-get`, `hx-post`, etc. for AJAX requests
- Implement proper error handling with HTMX responses
- Use out-of-band swaps for complex UI updates
- Handle HTMX-specific headers in handlers
- Leverage HTMX extensions sparingly and only when necessary

### JavaScript Guidelines
- **Location**: All custom JavaScript in `static/js/` directory
- **Purpose**: Only for functionality that HTMX cannot handle (e.g., color pickers, date pickers)
- **Size**: Keep functions small and focused on single responsibilities
- **Vanilla JS**: Avoid jQuery or other large libraries
- **Integration**: Use HTMX events and attributes to trigger custom JavaScript when needed

### Template Standards
- Use Templ for type-safe templates
- Templates in `{handler}/templ/` directories
- Shared templates in `src/shared/templ/`
- Follow consistent naming: `{feature}-{component}.templ`
- Prefer HTMX attributes over JavaScript event handlers

## Development Workflow

### Code Generation
- Run `templ generate` after template changes
- Use `//go:generate` directives in main.go
- Build CSS with `npm run build` before compilation

### Testing Standards
- Test files should end with `_test.go`
- Use table-driven tests where appropriate
- Mock external dependencies
- Test both success and error cases

### Environment Configuration
- Development: Local PostgreSQL with Docker Compose
- Production: Environment variables for configuration
- Use `.env` files for local development
- Never commit secrets to version control

## File Organization

### Directory Structure
```
src/
├── database/           # Database layer
│   ├── migrations/     # SQL migrations
│   └── repos/         # Repository implementations
├── handlers/          # HTTP handlers by domain
├── services/          # Business logic by domain
├── models/            # Domain models and forms
├── middleware/        # HTTP middleware
├── helpers/           # Utility functions
└── shared/            # Shared templates and utilities
```

### Naming Conventions
- Files: snake_case (e.g., `user_service.go`)
- Packages: lowercase, no underscores
- Functions: PascalCase for exported, camelCase for private
- Constants: ALL_CAPS for package-level constants

## Common Patterns to Follow

### Dependency Injection
```go
func SetupService(repoContainer *repos.RepoContainer) Service {
    return Service{
        repo: *repoContainer.GetRepo(),
        logger: slog.NewLogger("ServiceName"),
    }
}
```

### Error Handling
```go
if err != nil {
    s.logger.Error(err, "Operation failed - context: %s", context)
    return nil, errors.New("generic error message")
}
```

### Form Validation
```go
form := forms.NewForm()
if !isValid {
    form.SetError("field", "error message")
    return form, nil
}
```

## Performance Considerations

- Use database indexes appropriately
- Implement pagination for large datasets
- Cache static content
- Use connection pooling for database
- Minimize database queries in loops

## Deployment Standards

- Use Docker for containerization
- Environment-specific configuration
- Health check endpoints
- Proper logging for production debugging
- Database migration automation

## Common Commands

### Development
```bash
# Start local database
docker compose -f docker-compose-db-local.yml up -d

# Run migrations
docker run -v ./src/database/migrations:/migrations --network go-do-the-thing_default migrate/migrate -path=/migrations/ -database postgres://admin:admin@db:5432/todo_db?sslmode=disable up

# Generate templates
templ generate

# Build CSS
npm run build

# Run with hot reload
air
```

### Testing
```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

## Frontend Development Guidelines

### JavaScript Usage Policy
- **First Choice**: Use HTMX attributes for dynamic behavior
- **Second Choice**: Small vanilla JavaScript functions in `static/js/`
- **Avoid**: Large JavaScript frameworks, jQuery, or heavy libraries
- **Examples of Acceptable JS**: Color pickers, date pickers, form validation helpers
- **File Organization**: One file per feature (e.g., `colorpicker.js`, `datepicker.js`)

### Current JavaScript Files
- `colorpicker.js`: Custom color picker functionality
- `datepicker.js`: Date selection utilities  
- `functions.js`: General utility functions
- `sw.js`: Service worker for PWA features

### Adding New JavaScript
1. **Evaluate**: Can this be done with HTMX alone?
2. **Minimize**: Keep functions small and focused
3. **Vanilla**: Use pure JavaScript, no dependencies
4. **Integration**: Use HTMX events to trigger JavaScript when needed
5. **Testing**: Ensure functionality works without JavaScript (progressive enhancement)

### CSS and Styling
- **Primary**: Tailwind CSS utility classes
- **Custom**: `static/css/custom.css` for project-specific styles
- **Avoid**: CSS frameworks beyond Tailwind
- **Build Process**: `npm run build` compiles Tailwind

When working on this project, always:
1. **Prioritize HTMX**: Use HTMX for dynamic behavior before considering JavaScript
2. **Minimize Dependencies**: Avoid adding new JavaScript libraries
3. **Follow Established Patterns**: Use existing code structure and conventions
4. **Use Logging Consistently**: Implement proper structured logging
5. **Validate Input and Permissions**: Always check user authorization
6. **Handle Errors Appropriately**: Use custom error types and proper logging
7. **Write Tests**: Add tests for new functionality
8. **Update Migrations**: Include both UP and DOWN migration variants
9. **Use Middleware Stack**: Apply appropriate middleware for routes
10. **Follow Security Best Practices**: Never expose sensitive data or bypass auth