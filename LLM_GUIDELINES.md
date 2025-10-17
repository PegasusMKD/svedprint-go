# LLM Guidelines - Go Project Architecture

## Project Structure
```
project-root/
├── cmd/
│   └── api.go                    # Entry point only, minimal logic
├── internal/
│   ├── {domain}/                 # One per domain (student, school, etc.)
│   │   ├── dto.go                # API request/response structs
│   │   ├── handler.go            # Gin HTTP handlers
│   │   ├── mapper.go             # DTO ↔ domain model conversion
│   │   ├── repository.go         # Database access (wraps sqlc)
│   │   └── service.go            # Business logic
│   └── utility/                  # Shared helpers
├── db/
│   ├── migrations/               # golang-migrate files
│   ├── queries/                  # sqlc query definitions
│   └── sqlc/                     # Generated code (DO NOT EDIT)
├── go.mod
├── go.sum
└── sqlc.yaml
```

## Technology Stack
- **Database**: PostgreSQL with UUIDs as primary keys
- **Driver**: pgx/v5 (`github.com/jackc/pgx/v5`)
- **Query Generator**: sqlc (type-safe SQL)
- **Migrations**: golang-migrate (`github.com/golang-migrate/migrate/v4`)
- **HTTP Framework**: Gin (`github.com/gin-gonic/gin`)

## Layer Architecture

### Dependency Flow
```
HTTP Request → Handler → Mapper → Service → Repository → sqlc → PostgreSQL
```

### Layer Responsibilities

**handler.go** - `{Domain}Handler` struct
- Parse/validate HTTP requests (Gin)
- Call service methods
- Return HTTP responses with proper status codes
- Uses: DTOs, Service layer

**service.go** - `{Domain}Service` struct
- Business logic and validation
- Orchestrates repository calls
- Works with domain models (NOT DTOs)
- Uses: Repository layer

**repository.go** - `{Domain}Repository` struct
- Wraps sqlc generated queries
- Converts sqlc types ↔ domain models
- Handles DB errors
- Uses: sqlc.Queries

**dto.go** - Request/Response structs
- `{Domain}DTO`, `Create{Domain}Request`, `Update{Domain}Request`
- JSON tags for serialization
- Validation tags (binding)
- NEVER used in service/repository layers

**mapper.go** - Conversion functions
- `{Model}ToDTO(model) → DTO`
- `{Request}To{Model}(dto) → model`
- Handles type conversions (string ↔ pgtype.UUID, etc.)

## Naming Conventions

| Element | Convention | Example |
|---------|-----------|---------|
| Files | snake_case.go | student_service.go, dto.go |
| Packages | lowercase, single word | student, school |
| Structs | PascalCase | StudentHandler, SchoolService |
| Exported funcs | PascalCase | GetStudent, CreateSchool |
| Internal funcs | camelCase | parseRequest, validateInput |
| Variables | camelCase | studentID, schoolName |
| Constants | PascalCase or SCREAMING_SNAKE | MaxRetries, API_VERSION |

## Database Conventions

### Types (from pgx/v5)
- UUIDs: `pgtype.UUID`
- Nullable strings: `pgtype.Text`
- Dates: `pgtype.Date`
- Nullable ints: `pgtype.Int4`

### Schema Rules
- Primary keys: `uuid uuid primary key`
- Foreign keys: `references table(uuid)`
- Use PostgreSQL enums for constrained values
- Composite unique constraints where needed

### sqlc Patterns
**Query files** (`db/queries/{domain}.sql`):
```sql
-- name: GetStudentByUuid :one
select * from student where uuid = @student_uuid;

-- name: InsertStudent :one
insert into student (uuid, first_name, last_name)
values (gen_random_uuid(), @first_name, @last_name)
returning *;
```

**Configuration** (`sqlc.yaml`):
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

### Migrations
- Tool: `golang-migrate`
- Create: `migrate create -ext sql -dir db/migrations -seq {name}`
- Format: `YYYYMMDDHHMMSS_description.up.sql` and `.down.sql`
- ALWAYS create both up and down migrations

## Essential Code Pattern

**Complete layer flow example**:

```go
// handler.go
type StudentHandler struct {
    service *StudentService
}

func (h *StudentHandler) GetStudent(c *gin.Context) {
    uuid := c.Param("uuid")
    student, err := h.service.GetByUUID(c.Request.Context(), uuid)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
        return
    }
    c.JSON(http.StatusOK, StudentToDTO(student))
}

// service.go
type StudentService struct {
    repo *StudentRepository
}

func (s *StudentService) GetByUUID(ctx context.Context, uuid string) (*Student, error) {
    return s.repo.GetByUUID(ctx, uuid)
}

// repository.go
type StudentRepository struct {
    queries *sqlc.Queries
}

func (r *StudentRepository) GetByUUID(ctx context.Context, uuid string) (*Student, error) {
    pgUUID := pgtype.UUID{}
    if err := pgUUID.Scan(uuid); err != nil {
        return nil, fmt.Errorf("invalid uuid: %w", err)
    }
    return r.queries.GetStudentByUuid(ctx, pgUUID)
}
```

## Critical Guardrails

### DO
- Follow layered architecture: Handler → Service → Repository → sqlc
- Use dependency injection for all layers
- Pass `context.Context` as first parameter for I/O operations
- Handle errors explicitly at each layer
- Use DTOs only in handler layer
- Use domain models in service/repository layers
- Keep mappers separate and testable
- Use sqlc for all database queries
- Use golang-migrate for all migrations
- Use UUIDs for primary keys
- Define PostgreSQL enums for constrained values

### DO NOT
- Skip layers (e.g., handler calling repository directly)
- Put business logic in cmd/
- Let DTOs leak into service/repository layers
- Manually edit generated sqlc code (db/sqlc/*)
- Ignore errors
- Use `panic()` in library code
- Put SQL queries outside db/queries/
- Create migrations manually without golang-migrate
- Use string literals for enum values
- Mix database code with business logic

## Code Review Checklist
- [ ] Each domain has all 5 files: handler, service, repository, dto, mapper
- [ ] No layer skipping (handler → service → repository)
- [ ] DTOs isolated to handler layer
- [ ] Generated sqlc code is up to date
- [ ] Migrations have both .up.sql and .down.sql
- [ ] All errors handled at each layer
- [ ] Context propagated through call stack
- [ ] Naming conventions followed
- [ ] No business logic in cmd/
- [ ] Proper HTTP status codes in handlers

## Quick Commands
```bash
# Generate sqlc code
sqlc generate

# Create migration
migrate create -ext sql -dir db/migrations -seq migration_name

# Apply migrations
migrate -path db/migrations -database "postgresql://..." up
```
