# Go Project Guidelines and Guardrails

## Overview
This document outlines the architectural patterns, conventions, and guardrails for Go projects in this organization. These guidelines are designed to ensure consistency, maintainability, and scalability across codebases, particularly for rewrites and new projects.

---

## Project Structure

### Standard Layout
```
project-root/
├── cmd/                    # Application entry points
│   └── api.go             # Main application (e.g., API server)
├── internal/              # Private application code
│   ├── domain1/           # Domain-specific module
│   │   ├── dto.go         # Data Transfer Objects (API request/response)
│   │   ├── handler.go     # Gin HTTP handlers
│   │   ├── mapper.go      # Mapper between DTOs and domain models
│   │   ├── repository.go  # Database access layer (uses sqlc)
│   │   └── service.go     # Business logic
│   ├── domain2/
│   └── utility/           # Shared utilities
├── db/                    # Database layer
│   ├── migrations/        # SQL migration files (golang-migrate)
│   ├── queries/           # SQL query definitions (for sqlc)
│   └── sqlc/              # Generated code (DO NOT EDIT)
├── go.mod
├── go.sum
└── sqlc.yaml             # sqlc configuration
```

### Key Principles
1. **cmd/** - Contains only application entry points. Keep minimal logic here.
2. **internal/** - All business logic lives here. Code in internal/ cannot be imported by external projects.
3. **db/** - All database-related code, migrations, and queries are isolated here.

---

## Database Layer

### Technology Stack
- **Database**: PostgreSQL
- **Driver**: pgx/v5 (github.com/jackc/pgx/v5)
- **Query Builder**: sqlc (type-safe SQL query generation)
- **Migration Tool**: golang-migrate/migrate (github.com/golang-migrate/migrate)

### Database Organization

#### Migrations
- **Tool**: golang-migrate/migrate
- Location: `db/migrations/`
- Naming: `YYYYMMDDHHMMSS_description.up.sql` and `.down.sql`
- **Guardrails**:
  - ALWAYS create both up and down migrations
  - Use `migrate create` command to generate migration files
  - Use descriptive names for migrations
  - Keep migrations atomic and reversible
  - Define custom PostgreSQL types (ENUMs) at the top of the initial schema

**Creating Migrations**:
```bash
# Install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Create a new migration
migrate create -ext sql -dir db/migrations -seq migration_name

# Apply migrations
migrate -path db/migrations -database "postgresql://user:pass@localhost:5432/dbname?sslmode=disable" up

# Rollback last migration
migrate -path db/migrations -database "postgresql://user:pass@localhost:5432/dbname?sslmode=disable" down 1
```

#### Queries
- Location: `db/queries/`
- One file per domain/table (e.g., `school.sql`, `students.sql`)
- Use sqlc annotations for type-safe code generation
- **Guardrails**:
  - Always use named parameters with `@param_name` syntax
  - Specify query type (`:one`, `:many`, `:exec`)
  - Use descriptive query names (e.g., `GetSchoolByUuid`, `InsertSchool`)

**Example Query File** (`db/queries/school.sql`):
```sql
-- name: GetSchoolByUuid :one
select * from school
where uuid = @school_uuid;

-- name: InsertSchool :one
insert into school (uuid, school_name)
values (gen_random_uuid(), @school_name)
returning *;

-- name: UpdateSchool :one
update school
set director_name = @director_name
where uuid = @school_uuid
returning *;
```

#### sqlc Configuration
- Location: `sqlc.yaml` at project root
- **Standard Configuration**:
```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "db/queries"
    schema: "db/migrations"
    gen:
      go:
        package: "sqlc"
        out: "db/sqlc"
        sql_package: "pgx/v5"
```

#### Generated Code
- Location: `db/sqlc/`
- **Guardrails**:
  - NEVER manually edit generated code
  - Regenerate using `sqlc generate` after query changes
  - Generated files include:
    - `db.go` - Core database interface (DBTX, Queries)
    - `models.go` - Struct definitions for tables
    - `*.sql.go` - Query implementations

---

## Domain Layer Architecture (internal/)

### Organization
Each domain follows a layered architecture with clear separation of concerns:

```
internal/
├── student/
│   ├── dto.go           # Data Transfer Objects (DTOs)
│   ├── handler.go       # HTTP handlers (Gin)
│   ├── mapper.go        # DTO ↔ Model mapping
│   ├── repository.go    # Database access layer
│   └── service.go       # Business logic
├── school/
│   ├── dto.go
│   ├── handler.go
│   ├── mapper.go
│   ├── repository.go
│   └── service.go
└── utility/
    └── utilities.go     # Shared helper functions
```

### Layer Responsibilities

#### 1. Handler Layer (`handler.go`)
- **Purpose**: HTTP request/response handling using Gin framework
- **Responsibilities**:
  - Parse and validate incoming requests
  - Call service layer methods
  - Format and return HTTP responses
  - Handle HTTP-specific errors (400, 404, 500, etc.)
- **Dependencies**: Service layer, DTOs
- **Naming**: `{Domain}Handler` struct with methods like `GetStudent`, `CreateStudent`

**Example**:
```go
type StudentHandler struct {
    service *StudentService
}

func (h *StudentHandler) GetStudent(c *gin.Context) {
    uuid := c.Param("uuid")
    student, err := h.service.GetStudentByUUID(c.Request.Context(), uuid)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
        return
    }
    dto := MapStudentToDTO(student)
    c.JSON(http.StatusOK, dto)
}
```

#### 2. Service Layer (`service.go`)
- **Purpose**: Business logic and orchestration
- **Responsibilities**:
  - Implement business rules and validation
  - Coordinate between multiple repositories if needed
  - Handle domain-specific logic
  - Return domain models (not DTOs)
- **Dependencies**: Repository layer
- **Naming**: `{Domain}Service` struct

**Example**:
```go
type StudentService struct {
    repo *StudentRepository
}

func (s *StudentService) GetStudentByUUID(ctx context.Context, uuid string) (*Student, error) {
    // Business logic here
    return s.repo.GetByUUID(ctx, uuid)
}
```

#### 3. Repository Layer (`repository.go`)
- **Purpose**: Database access abstraction
- **Responsibilities**:
  - Wrap sqlc generated queries
  - Convert between sqlc types and domain models
  - Handle database-specific errors
  - Provide a clean interface for the service layer
- **Dependencies**: sqlc generated code
- **Naming**: `{Domain}Repository` struct

**Example**:
```go
type StudentRepository struct {
    queries *sqlc.Queries
}

func (r *StudentRepository) GetByUUID(ctx context.Context, uuid string) (*Student, error) {
    pgUUID := pgtype.UUID{}
    err := pgUUID.Scan(uuid)
    if err != nil {
        return nil, fmt.Errorf("invalid uuid: %w", err)
    }

    sqlcStudent, err := r.queries.GetStudentByUuid(ctx, pgUUID)
    if err != nil {
        return nil, err
    }

    return fromSQLCStudent(sqlcStudent), nil
}
```

#### 4. DTO Layer (`dto.go`)
- **Purpose**: Define API contracts (request/response structures)
- **Responsibilities**:
  - Represent external API structure
  - Include JSON tags for serialization
  - Validation tags (if using validator)
  - Keep separate from internal domain models
- **Naming**: `{Domain}DTO`, `Create{Domain}Request`, `Update{Domain}Request`

**Example**:
```go
type StudentDTO struct {
    UUID        string `json:"uuid"`
    FirstName   string `json:"first_name"`
    LastName    string `json:"last_name"`
    DateOfBirth string `json:"date_of_birth,omitempty"`
}

type CreateStudentRequest struct {
    FirstName   string `json:"first_name" binding:"required"`
    LastName    string `json:"last_name" binding:"required"`
    DateOfBirth string `json:"date_of_birth,omitempty"`
}
```

#### 5. Mapper Layer (`mapper.go`)
- **Purpose**: Convert between DTOs and domain models
- **Responsibilities**:
  - DTO → Domain model conversion
  - Domain model → DTO conversion
  - Handle type conversions (e.g., string UUID ↔ pgtype.UUID)
  - Keep mapping logic separate and testable
- **Naming**: `{DomainModel}ToDTO`, `DTOTo{DomainModel}`

**Example**:
```go
func StudentToDTO(s *Student) *StudentDTO {
    return &StudentDTO{
        UUID:        s.UUID.String(),
        FirstName:   s.FirstName,
        LastName:    s.LastName,
        DateOfBirth: s.DateOfBirth.Format("2006-01-02"),
    }
}

func CreateRequestToStudent(req *CreateStudentRequest) *Student {
    return &Student{
        FirstName: req.FirstName,
        LastName:  req.LastName,
        // ... other fields
    }
}
```

### Naming Conventions
- **Files**: `{layer}.go` (e.g., `handler.go`, `service.go`, `dto.go`)
- **Package names**: Match directory names (single word, lowercase)
- **Struct names**: `{Domain}{Layer}` (e.g., `StudentHandler`, `StudentService`)

### Dependency Flow
```
HTTP Request
    ↓
Handler (Gin) → validates, parses DTOs
    ↓
Mapper → converts DTO to domain model
    ↓
Service → business logic
    ↓
Repository → database access
    ↓
sqlc Queries → type-safe SQL
    ↓
PostgreSQL
```

### Guardrails
- **Separation of Concerns**: Each layer has a single responsibility
- **Dependency Direction**: Handler → Service → Repository → sqlc
- **No Layer Skipping**: Handlers must not call repositories directly
- **DTO Isolation**: DTOs never leak into service or repository layers
- **Domain Models**: Service and repository work with domain models, not DTOs
- **Error Handling**: Each layer handles its own error types appropriately

---

## Code Conventions

### Naming
- **Files**: `snake_case.go` (e.g., `student_service.go`)
- **Packages**: Single word, lowercase, descriptive (e.g., `student`, `school`)
- **Types**: PascalCase (e.g., `StudentService`, `School`)
- **Functions/Methods**: PascalCase for exported, camelCase for internal
- **Variables**: camelCase
- **Constants**: PascalCase or SCREAMING_SNAKE_CASE for global constants

### Database Types
- Use `pgtype.UUID` for UUIDs (from pgx/v5)
- Use `pgtype.Text` for nullable strings
- Use `pgtype.Date` for dates
- Use `pgtype.Int4` for nullable integers

### Error Handling
- Always check errors
- Return errors up the call stack
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Use `errors.Is()` and `errors.As()` for error checking

### Context
- Always pass `context.Context` as the first parameter to functions that perform I/O
- Use `context.Background()` at entry points
- Propagate context through the call stack

---

## PostgreSQL Schema Design

### Best Practices
1. **Primary Keys**: Use UUIDs (`uuid primary key`)
2. **Foreign Keys**: Always define relationships with `references`
3. **Enums**: Define custom types for constrained values
4. **Defaults**: Set sensible defaults where appropriate
5. **Constraints**: Use unique constraints and check constraints

**Example Schema Patterns**:
```sql
-- Custom types
create type gender as enum ('male', 'female', 'unknown');
create type academic_level as enum ('first_year', 'second_year', 'junior_year', 'senior_year');

-- Table with UUID primary key
create table school (
    uuid uuid primary key,
    school_name text not null unique,
    director_name text,
    created_at timestamp not null default now()
);

-- Table with foreign keys and defaults
create table student (
    uuid uuid primary key,
    first_name text not null,
    last_name text not null,
    school_uuid uuid not null references school(uuid),
    gender gender not null default 'unknown',
    created_at timestamp not null default now()
);

-- Composite unique constraints
create table students_yearly_detail (
    uuid uuid primary key,
    student_uuid uuid not null references student(uuid),
    academic_level academic_level not null,
    constraint uq_student_yearly_detail unique (student_uuid, academic_level)
);
```

---

## Dependency Management

### go.mod Guidelines
- Use Go 1.21+ for modern features
- Keep dependencies minimal and well-maintained
- Pin major versions explicitly
- Use `go mod tidy` regularly

### Required Dependencies
- `github.com/jackc/pgx/v5` - PostgreSQL driver
- `github.com/gin-gonic/gin` - HTTP web framework
- `github.com/golang-migrate/migrate/v4` - Database migrations
- Additional dependencies as needed per project

---

## Testing Standards

### Organization
- Test files: `*_test.go` in the same package
- Integration tests: `tests/` or `*_integration_test.go`
- Use table-driven tests where appropriate

### Testing Database
- Use a separate test database or docker containers
- Clean up test data after each test
- Use transactions and rollback for isolation

---

## Git Workflow

### Branches
- `master` or `main` - production-ready code
- Feature branches: `feature/description`
- Bug fixes: `fix/description`

### Commits
- Use descriptive commit messages
- Format: `type: brief description`
- Types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`

---

## Migration from Other Languages (e.g., Python to Go)

### When Rewriting to Go
1. **Analyze the existing structure** - Understand domain boundaries
2. **Start with the database layer** - Define schema and migrations first
3. **Generate sqlc code** - Set up queries before business logic
4. **Build services incrementally** - One domain at a time
5. **Maintain API compatibility** - If rewriting an API, keep endpoints consistent

### Python to Go Mapping
- **Python classes** → Go structs with methods
- **Python modules** → Go packages
- **Python ORM (SQLAlchemy, etc.)** → sqlc with pgx
- **Python type hints** → Go static types
- **Python decorators** → Go middleware/wrapper functions

---

## Development Workflow

### Setup New Project
1. Initialize Go module: `go mod init github.com/org/project`
2. Create directory structure (cmd/, internal/, db/)
3. Set up database migrations
4. Configure sqlc.yaml
5. Write SQL queries in db/queries/
6. Generate code: `sqlc generate`
7. Implement services in internal/

### Running sqlc
```bash
# Generate Go code from SQL
sqlc generate

# Verify configuration
sqlc verify
```

### Database Migrations with golang-migrate
```bash
# Create a new migration
migrate create -ext sql -dir db/migrations -seq migration_name

# Apply all pending migrations
migrate -path db/migrations -database "postgresql://user:pass@localhost:5432/dbname?sslmode=disable" up

# Rollback the last migration
migrate -path db/migrations -database "postgresql://user:pass@localhost:5432/dbname?sslmode=disable" down 1

# Check migration version
migrate -path db/migrations -database "postgresql://user:pass@localhost:5432/dbname?sslmode=disable" version
```

---

## Common Pitfalls and Guardrails

### DO
- Use sqlc for type-safe database queries
- Keep business logic in internal/ services
- Use dependency injection for all layers
- Write tests for critical paths
- Use context.Context for cancellation
- Handle errors explicitly at each layer
- Use UUIDs for primary keys
- Define database enums for constrained values
- Use golang-migrate for all database migrations
- Follow the layered architecture (Handler → Service → Repository)
- Use DTOs for API contracts, domain models internally
- Keep mappers separate and testable
- Use Gin for HTTP routing and middleware

### DO NOT
- Put business logic in cmd/
- Manually edit generated sqlc code
- Ignore errors
- Use `panic()` in library code (only in main)
- Mix database code with business logic
- Skip database migrations
- Use string literals for enum values (use generated types)
- Expose internal/ packages to external projects
- Let DTOs leak into service or repository layers
- Skip layers (e.g., handler calling repository directly)
- Put SQL queries outside of db/queries/
- Create migrations manually without golang-migrate

---

## Code Review Checklist

Before submitting code:
- [ ] Generated sqlc code is up to date
- [ ] All errors are handled at each layer
- [ ] Tests pass
- [ ] Migration files (up and down) are present
- [ ] Migrations created using golang-migrate
- [ ] No business logic in cmd/
- [ ] Each domain has all required layers (handler, service, repository, dto, mapper)
- [ ] Layered architecture is respected (no layer skipping)
- [ ] DTOs are isolated to handler layer
- [ ] Mappers are used for all DTO ↔ model conversions
- [ ] Services follow single responsibility principle
- [ ] Context is propagated correctly
- [ ] Database queries use named parameters
- [ ] Code follows naming conventions
- [ ] Dependencies are minimal and necessary
- [ ] Gin handlers use proper HTTP status codes

---

## Resources

- [sqlc Documentation](https://docs.sqlc.dev/)
- [pgx Documentation](https://pkg.go.dev/github.com/jackc/pgx/v5)
- [Gin Documentation](https://gin-gonic.com/docs/)
- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Effective Go](https://go.dev/doc/effective_go)

---

## Questions or Clarifications

For questions about these guidelines or suggestions for improvements, please:
1. Open an issue in the project repository
2. Discuss in team meetings
3. Update this document with consensus changes

**Last Updated**: 2025-10-17
