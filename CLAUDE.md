# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a rewrite of a student management and diploma printing system (svedprint) from Python to Go. The application manages schools, students, teachers, academic years, subjects, and generates diplomas/testimonies for students.

## Essential Commands

### Database Operations
```bash
# Generate sqlc code after modifying queries
sqlc generate

# Create new migration
migrate create -ext sql -dir db/migrations -seq migration_name

# Apply migrations (replace with actual DB connection string)
migrate -path db/migrations -database "postgresql://user:pass@localhost:5432/dbname?sslmode=disable" up

# Rollback last migration
migrate -path db/migrations -database "postgresql://user:pass@localhost:5432/dbname?sslmode=disable" down 1
```

### Development Workflow
```bash
# Build and run the API
go run cmd/api.go

# Run tests (when implemented)
go test ./...

# Run tests for specific package
go test ./internal/school/...

# Update dependencies
go mod tidy
```

## Architecture

### Layered Architecture Pattern

The codebase follows a strict layered architecture: **Handler → Mapper → Service → Repository → sqlc**

**Critical Rule**: Never skip layers. Handlers must call Services, Services must call Repositories.

### Domain Structure

Each domain (student, school, teacher, subject, etc.) follows this pattern:
```
internal/{domain}/
├── dto.go          # API request/response structs (JSON tags, validation)
├── handler.go      # Gin HTTP handlers (parse requests, return responses)
├── mapper.go       # DTO ↔ domain model conversion
├── repository.go   # Database access layer (wraps sqlc queries)
└── service.go      # Business logic and orchestration
```

**Current State**: Most domain folders only contain empty `{domain}_service.go` files. The implementation is in early stages.

### Database Layer

- **Engine**: PostgreSQL with pgx/v5 driver
- **Query Generation**: sqlc (type-safe SQL-to-Go code generation)
- **Migrations**: golang-migrate
- **Schema Location**: `db/migrations/20251006185427_initial_schema.up.sql`

Key schema entities:
- `school` - School information (director, location, business details)
- `student` - Student personal information
- `students_yearly_detail` - Academic records per year/level
- `teacher` - Teacher accounts with print permissions
- `subject` and `subject_package` - Course management
- `academic_year` - Year configuration per school type (gymnasium/professional)
- `school_class` - Class organization with responsible teachers

**Important Types**:
- UUIDs: `pgtype.UUID` (all primary keys)
- Nullable strings: `pgtype.Text`
- Dates: `pgtype.Date`
- PostgreSQL enums: `gender`, `academic_level`, `school_type`, `study_type`, `behaviour_type`, `year_success_type`

### Query Files (`db/queries/`)

One SQL file per domain entity containing sqlc-annotated queries:
- `school.sql` - GetSchoolByUuid, InsertSchool, UpdateSchool
- `students.sql`, `teacher.sql`, `subject.sql`, etc.

All queries use named parameters (`@param_name`) and specify return type (`:one`, `:many`, `:exec`).

## Critical Patterns

### DTO vs Domain Model Separation

**DTOs** (`dto.go`) - Used ONLY in handler layer:
- JSON tags for serialization
- Validation tags (binding)
- External API representation

**Domain Models** - Used in service/repository layers:
- Match database structure (often sqlc-generated)
- Internal business logic representation

### Mapper Pattern

Convert between DTOs and domain models in `mapper.go`:
```go
func StudentToDTO(s *Student) *StudentDTO { ... }
func CreateRequestToStudent(req *CreateStudentRequest) *Student { ... }
```

### Error Handling

- Propagate `context.Context` as first parameter for all I/O operations
- Handle errors at each layer with appropriate context
- Handlers return proper HTTP status codes (404, 400, 500, etc.)

## Important Constraints

### DO
- Follow Handler → Service → Repository → sqlc flow
- Use dependency injection for all layers
- Keep DTOs isolated to handler layer only
- Use sqlc for ALL database queries
- Use golang-migrate for ALL schema changes
- Use UUIDs for primary keys (via `gen_random_uuid()`)
- Respect PostgreSQL enums defined in schema
- Pass `context.Context` through call stack
- Keep business logic in service layer, NOT in cmd/

### DO NOT
- Skip layers (e.g., handler calling repository directly)
- Let DTOs leak into service/repository layers
- Manually edit `db/sqlc/*` generated code
- Put business logic in `cmd/api.go`
- Create migrations manually without golang-migrate
- Put SQL queries outside `db/queries/`
- Use string literals for enum values
- Ignore errors or use panic() in library code

## Domain-Specific Context

### Academic Structure
- **Academic levels**: first_year, second_year, junior_year, senior_year
- **School types**: gymnasium (general education), professional (vocational)
- **Study types**: regular, irregular
- Students can have different subject packages (general, pma, pmb orientations)
- Composite unique constraint on students_yearly_detail prevents duplicate records per student/year/level/class

### Key Business Relationships
- Students belong to one school but have yearly details per academic_level
- Each yearly detail links to: school_class, academic_year, subject_package
- Teachers can be assigned as responsible for school_class
- Teachers have print permissions controlled by `print_allowed` flag
- Academic years are per school and school_type, with act_number/act_date for official records

## Technology Stack
- Go 1.24+
- PostgreSQL (pgx/v5 driver)
- sqlc v1.29.0+
- golang-migrate
- Gin web framework (planned, not yet implemented)

## Configuration

### sqlc.yaml
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

## Reference Documentation

See `LLM_GUIDELINES.md` and `GUIDELINES.md` for complete architectural patterns, naming conventions, and code examples.
