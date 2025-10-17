# Day 1: Go Project Initialization

**Phase:** 1 - Project Setup & Database Foundation
**Day:** 1 of 50
**Focus:** Initialize Go project, install dependencies, create basic structure

---

## Goals for Today

- [X] Initialize Go module and verify Go version
- [ ] Install all required dependencies
- [ ] Create project directory structure
- [ ] Set up version control (.gitignore)
- [ ] Install development tools (sqlc, migrate)

---

## Tasks

### Morning Session (4 hours)

#### 1. Environment Setup (30 min)
- [X] Verify Go 1.21+ is installed
  ```bash
  go version
  ```
- [X] Navigate to project directory
  ```bash
  cd /path/to/svedprint-go
  ```
- [X] Initialize Git repository (if not already done)
  ```bash
  git init
  ```

#### 2. Go Module Initialization (15 min)
- [X] Initialize Go module
  ```bash
  go mod init github.com/pazzio/svedprint
  ```
- [X] Verify go.mod file created
- [X] Set Go version in go.mod to 1.21+

#### 3. Install Core Dependencies (1 hour)
- [X] Install Gin framework
  ```bash
  go get github.com/gin-gonic/gin
  ```
- [X] Install pgx/v5 (PostgreSQL driver)
  ```bash
  go get github.com/jackc/pgx/v5
  go get github.com/jackc/pgx/v5/stdlib
  ```
- [X] Install golang-migrate library
  ```bash
  go get github.com/golang-migrate/migrate/v4
  ```
- [X] Install validator
  ```bash
  go get github.com/go-playground/validator/v10
  ```
- [X] Install Keycloak dependencies
  ```bash
  go get github.com/Nerzal/gocloak/v13
  go get github.com/golang-jwt/jwt/v5
  go get github.com/MicahParks/keyfunc/v2
  ```
- [X] Install testing dependencies
  ```bash
  go get github.com/stretchr/testify
  go get github.com/DATA-DOG/go-sqlmock
  ```
- [X] Run go mod tidy
  ```bash
  go mod tidy
  ```
- [X] Verify all dependencies downloaded (check go.sum)

#### 4. Install Development Tools (45 min)
- [X] Install sqlc globally
  ```bash
  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
  ```
- [X] Verify sqlc installation
  ```bash
  sqlc version
  ```
- [X] Install golang-migrate CLI
  ```bash
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```
- [X] Verify migrate installation
  ```bash
  migrate -version
  ```
- [X] Add GOPATH/bin to PATH if needed
  ```bash
  export PATH=$PATH:$(go env GOPATH)/bin
  ```

### Afternoon Session (4 hours)

#### 5. Create Project Structure (1 hour)
- [X] Create directory structure:
  ```bash
  mkdir -p cmd
  mkdir -p internal/utility
  mkdir -p db/migrations
  mkdir -p db/queries
  mkdir -p db/sqlc
  mkdir -p scripts/backup
  mkdir -p docs
  mkdir -p backups
  ```
- [X] Create placeholder files:
  ```bash
  touch cmd/api.go
  touch internal/utility/.gitkeep
  touch db/queries/.gitkeep
  ```
- [X] Verify directory structure
  ```bash
  tree -L 2
  ```

#### 6. Create .gitignore (30 min)
- [X] Create comprehensive .gitignore file with:
  ```gitignore
  # Binaries
  *.exe
  *.exe~
  *.dll
  *.so
  *.dylib
  /svedprint-api
  /svedprint

  # Test binary, built with `go test -c`
  *.test

  # Output of the go coverage tool
  *.out
  coverage.out
  coverage.html

  # Generated code
  db/sqlc/*
  !db/sqlc/.gitkeep

  # Environment files
  .env
  .env.local
  .env.*.local

  # IDE
  .vscode/
  .idea/
  *.swp
  *.swo
  *~

  # Logs
  *.log
  logs/

  # Backups
  backups/*.sql
  backups/*.dump
  !backups/.gitkeep

  # OS
  .DS_Store
  Thumbs.db

  # Temporary files
  tmp/
  temp/
  ```
- [X] Create .gitkeep files for empty directories:
  ```bash
  touch db/sqlc/.gitkeep
  touch backups/.gitkeep
  ```

#### 7. Create Initial README (1 hour)
- [X] Create README.md with:
  - [X] Project title and description
  - [X] Prerequisites (Go 1.21+, PostgreSQL 14+, Docker)
  - [X] Quick start instructions (placeholder)
  - [X] Project structure explanation
  - [X] Link to detailed docs
  - [X] License information
- [X] Example structure:
  ```markdown
  # Svedprint - School Certificate Management System

  Go backend for managing school certificates and diplomas.

  ## Prerequisites
  - Go 1.21+
  - PostgreSQL 14+
  - Docker & Docker Compose
  - Keycloak 21+

  ## Project Structure
  ```
  svedprint-go/
  ├── cmd/              # Application entry points
  ├── internal/         # Private application code
  ├── db/               # Database migrations and queries
  ├── scripts/          # Utility scripts
  └── docs/             # Documentation
  ```

  ## Getting Started
  (To be filled in as implementation progresses)
  ```

#### 8. Create .env.example (30 min)
- [X] Create .env.example with all required environment variables:
  ```env
  # Database Configuration
  DATABASE_URL=postgresql://user:password@localhost:5432/svedprint_db?sslmode=disable
  DATABASE_MAX_CONNS=25
  DATABASE_MAX_IDLE_CONNS=10
  DATABASE_CONN_MAX_LIFETIME=5m

  # Keycloak Configuration
  KEYCLOAK_URL=http://localhost:8080
  KEYCLOAK_REALM=svedprint
  KEYCLOAK_CLIENT_ID=svedprint-backend
  KEYCLOAK_CLIENT_SECRET=your-client-secret-here
  KEYCLOAK_JWKS_URL=http://localhost:8080/realms/svedprint/protocol/openid-connect/certs

  # Application Configuration
  PORT=8000
  GIN_MODE=debug
  LOG_LEVEL=debug

  # Security
  JWT_SECRET=your-jwt-secret-here
  CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8000
  ```
- [X] Add note in README about copying .env.example to .env

#### 9. Initial Git Commit (30 min)
- [X] Stage all files
  ```bash
  git add .
  ```
- [X] Review what will be committed
  ```bash
  git status
  ```
- [X] Create initial commit
  ```bash
  git commit -m "Initial project setup

  - Initialize Go module
  - Add all dependencies (Gin, pgx, Keycloak, etc.)
  - Create project directory structure
  - Add .gitignore and .env.example
  - Add initial README
  "
  ```
- [X] Verify commit
  ```bash
  git log --oneline
  ```

---

## Testing

### Verification Steps
- [X] Run `go mod verify` - should pass
- [X] Run `go build ./...` - should compile without errors
- [X] Run `sqlc version` - should show version number
- [X] Run `migrate -version` - should show version number
- [X] Verify all directories created: `ls -la`

---

## Documentation

- [x] Created README.md with basic project info
- [x] Created .env.example with all configuration
- [X] Document any issues encountered
- [X] Update project-implementation-plan with actual progress

---

## Blockers & Issues

**Potential Issues:**
- Go version too old (need 1.21+)
- GOPATH not set correctly
- migrate or sqlc installation fails
- Permission issues creating directories

**Solutions:**
- Install/upgrade Go from official website
- Set GOPATH: `export GOPATH=$HOME/go`
- Check $GOPATH/bin is in PATH
- Use sudo for system-wide tool installation

---

## Tomorrow's Preview

**Day 2 Focus:**
- Set up PostgreSQL database (dev and test)
- Configure sqlc
- Create first database migrations (enums, school, academic_year)
- Verify migrations work

**Preparation:**
- Ensure PostgreSQL is installed and running
- Have database admin credentials ready
- Review database schema design in DJANGO_TO_GO_MAPPING.md

---

## Notes

- Keep dependencies up to date but use stable versions
- All development tools installed globally for convenience
- Project structure follows Go best practices
- .gitignore prevents committing generated code
- .env.example serves as documentation for required config

---

## Time Tracking

**Estimated:** 8 hours
**Actual:** 2 hours
**Difference:** 6 hours

**Completed?** [X] Yes [ ] No
**Blockers:** None
**Notes:** None
