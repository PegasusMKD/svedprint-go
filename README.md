# Svedprint - Student Management & Diploma Printing System

A microservices-based student management and diploma printing system built with Go, designed for deployment on Railway.

## Architecture Overview

This project implements a modern microservices architecture with the following components:

### Services

1. **Gateway** (Port 8000) - Public-facing
   - JWT authentication via Keycloak
   - Reverse proxy to internal services
   - Request logging with worker pool
   - Database: `gateway_db`

2. **Svedprint** (Port 8001) - Private
   - Main business logic (students, schools, teachers, grades)
   - Redis caching with cache-aside pattern
   - Database: `svedprint_db`

3. **Svedprint Admin** (Port 8002) - Private
   - Administrative operations
   - Data migration management
   - Audit logging
   - Database: `admin_db`

4. **Svedprint Print** (Port 8003) - Private
   - Stateless PDF/image generation
   - No database required

### Infrastructure

- **PostgreSQL**: Database-per-service pattern (3 databases + Keycloak)
- **Redis**: Caching layer for read-heavy data
- **Keycloak**: OAuth2/OIDC authentication provider

## Project Structure

```
svedprint-go/
├── cmd/                          # Service entry points
│   ├── gateway/
│   │   └── main.go              # Gateway service
│   ├── svedprint/
│   │   └── main.go              # Main service
│   ├── svedprint-admin/
│   │   └── main.go              # Admin service
│   └── svedprint-print/
│       └── main.go              # Print service
│
├── internal/                     # Private application code
│   ├── gateway/                 # Gateway implementation
│   ├── svedprint/               # Main service implementation
│   ├── svedprint-admin/         # Admin service implementation
│   └── svedprint-print/         # Print service implementation
│
├── pkg/                         # Shared packages
│   ├── config/                  # Configuration loader
│   ├── database/                # Database helpers & migrations
│   ├── jwt/                     # JWT validation
│   ├── logger/                  # Logging setup
│   └── redis/                   # Redis client wrapper
│
├── db/                          # Database files (per service)
│   ├── gateway/
│   │   ├── migrations/          # Gateway DB migrations
│   │   └── queries/             # Gateway SQL queries
│   ├── svedprint/
│   │   ├── migrations/          # Main DB migrations
│   │   └── queries/             # Main SQL queries
│   └── svedprint-admin/
│       ├── migrations/          # Admin DB migrations
│       └── queries/             # Admin SQL queries
│
├── scripts/
│   └── init-db.sh              # PostgreSQL multi-database setup
│
├── .env.example                 # Environment variable template
├── docker-compose.yml           # Local development setup
├── Dockerfile                   # Multi-stage build for all services
├── railway.json                 # Railway deployment config
│
├── gateway-sqlc.yaml           # Gateway sqlc configuration
├── svedprint-sqlc.yaml         # Main service sqlc configuration
├── svedprint-admin-sqlc.yaml   # Admin service sqlc configuration
│
├── DOCKER_SETUP.md             # Local development guide
├── RAILWAY_DEPLOYMENT.md       # Production deployment guide
├── CLAUDE.md                   # Claude Code context
└── README.md                   # This file
```

## Key Design Principles

### 1. Database-per-Service
Each service owns its database. Services never share databases directly.

```
gateway       → gateway_db
svedprint     → svedprint_db
admin         → admin_db
print         → (stateless, no database)
```

### 2. Service Communication
- **Gateway**: Only public-facing service
- **Internal services**: Private, accessible only via gateway
- **Data access**: Services call each other via internal APIs (no shared databases)

### 3. Authentication Flow
```
Client → Gateway (validates JWT) → Internal Service (trusts gateway)
                ↑
          Keycloak JWKS
```

### 4. Request Logging
Gateway uses a worker pool pattern for non-blocking request logging:
- Gin middleware adds logs to buffered channel
- Worker pool (5 goroutines) processes logs in background
- Batch inserts into `gateway_db`

### 5. Caching Strategy
Svedprint service implements cache-aside pattern with Redis:
- Check cache first
- On miss: fetch from DB, populate cache
- TTL-based expiration

## Quick Start

### Local Development (Docker Compose)

```bash
# 1. Copy environment file
cp .env.example .env

# 2. Edit .env with your configuration
# (At minimum, set KEYCLOAK_CLIENT_SECRET)

# 3. Start all services
docker-compose up -d

# 4. Check status
docker-compose ps

# 5. View logs
docker-compose logs -f gateway
```

See [DOCKER_SETUP.md](./DOCKER_SETUP.md) for detailed instructions.

### Production Deployment (Railway)

See [RAILWAY_DEPLOYMENT.md](./RAILWAY_DEPLOYMENT.md) for step-by-step Railway deployment guide.

## Development Workflow

### 1. Install Dependencies

```bash
go mod download
```

### 2. Generate sqlc Code

After modifying SQL queries:

```bash
sqlc generate -f svedprint-sqlc.yaml
sqlc generate -f svedprint-admin-sqlc.yaml
sqlc generate -f gateway-sqlc.yaml
```

### 3. Create Migration

```bash
# Example for svedprint service
migrate create -ext sql -dir db/svedprint/migrations -seq add_new_table
```

### 4. Run Locally (without Docker)

```bash
# Set environment variables
export DATABASE_URL="postgresql://..."
export PORT="8001"

# Run service
go run cmd/svedprint/main.go
```

### 5. Build Binary

```bash
# Build specific service
go build -o bin/svedprint ./cmd/svedprint

# Build all services
go build -o bin/gateway ./cmd/gateway
go build -o bin/svedprint ./cmd/svedprint
go build -o bin/svedprint-admin ./cmd/svedprint-admin
go build -o bin/svedprint-print ./cmd/svedprint-print
```

## Technology Stack

- **Language**: Go 1.24+
- **Web Framework**: Gin
- **Database**: PostgreSQL 16 (via pgx/v5)
- **Cache**: Redis 7
- **Auth**: Keycloak 23
- **Query Builder**: sqlc
- **Migrations**: golang-migrate
- **Logging**: zerolog
- **JWT**: golang-jwt/jwt

## Environment Variables

Key environment variables (see `.env.example` for complete list):

| Variable | Description | Required |
|----------|-------------|----------|
| `PORT` | Service port | Yes |
| `DATABASE_URL` | PostgreSQL connection string | Yes (except print) |
| `GATEWAY_DATABASE_URL` | Gateway DB connection | Yes (gateway only) |
| `KEYCLOAK_JWKS_URL` | Keycloak public keys URL | Yes (gateway only) |
| `REDIS_ADDR` | Redis address | Yes (svedprint only) |
| `SVEDPRINT_SERVICE_URL` | Main service URL | Yes (admin & print) |
| `GIN_MODE` | Gin mode (debug/release) | No |
| `LOG_LEVEL` | Log level | No |

## API Structure

### Gateway Routes
```
GET  /health                    # Health check
POST /api/auth/login           # Proxy to Keycloak
GET  /api/svedprint/*          # Proxy to svedprint service
GET  /api/admin/*              # Proxy to admin service
POST /api/print/*              # Proxy to print service
```

### Authentication
All requests (except `/health`) require a valid JWT token:

```bash
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8000/api/svedprint/schools
```

## Database Schema

### Svedprint DB
- `school` - School information
- `student` - Student records
- `teacher` - Teacher accounts
- `academic_year` - Year configurations
- `subject`, `subject_package` - Course management
- `students_yearly_detail` - Academic records per year
- `school_class` - Class organization

### Admin DB
- `admin_users` - Admin user accounts
- `admin_audit_logs` - Audit trail
- `data_migrations` - Migration tracking
- `system_config` - System configuration

### Gateway DB
- `request_logs` - HTTP request logs

## Monitoring & Debugging

### Health Checks

Each service exposes a `/health` endpoint:

```bash
curl http://localhost:8000/health  # Gateway
curl http://localhost:8001/health  # Svedprint
curl http://localhost:8002/health  # Admin
curl http://localhost:8003/health  # Print
```

### Logs

**Docker Compose**:
```bash
docker-compose logs -f <service-name>
```

**Railway**:
```bash
railway logs --service=<service-name>
```

### Database Access

**Local (Docker)**:
```bash
docker-compose exec postgres psql -U postgres -d svedprint_db
```

**Railway**:
```bash
railway run psql $DATABASE_URL
```

## Testing

```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./internal/svedprint/...

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...
```

## Security Considerations

1. **Authentication**: All requests validated by gateway via Keycloak JWT
2. **Internal Services**: Not exposed publicly, only accessible via gateway
3. **Database Isolation**: Each service has its own database
4. **Secret Management**: Use Railway's secret variables in production
5. **HTTPS**: Use Railway's automatic HTTPS for public domains
6. **Rate Limiting**: Implement in gateway (TODO)
7. **CORS**: Configure in gateway for frontend origins

## Contributing

1. Follow the layered architecture: Handler → Service → Repository → sqlc
2. Use sqlc for all database queries
3. Use golang-migrate for schema changes
4. Write tests for business logic
5. Follow Go best practices and conventions
6. See [GUIDELINES.md](./GUIDELINES.md) for detailed patterns

## License

[Add your license here]

## Support

For issues or questions:
- Check [DOCKER_SETUP.md](./DOCKER_SETUP.md) for local development
- Check [RAILWAY_DEPLOYMENT.md](./RAILWAY_DEPLOYMENT.md) for deployment
- Review [CLAUDE.md](./CLAUDE.md) for architecture details
- Open an issue in the repository
