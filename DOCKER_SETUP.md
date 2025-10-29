# Docker Compose Local Development Setup

This guide explains how to run the entire svedprint microservices stack locally using Docker Compose.

## Prerequisites

- Docker 20.10+
- Docker Compose v2.0+
- Git
- Make (optional, for convenience commands)

## Quick Start

1. **Clone the repository** (if not already done):
   ```bash
   git clone <your-repo-url>
   cd svedprint-go
   ```

2. **Copy the environment file**:
   ```bash
   cp .env.example .env
   ```

3. **Edit `.env` file** with your configuration:
   ```bash
   # Minimal required changes:
   # - Set KEYCLOAK_CLIENT_SECRET (generate a secure random string)
   # - Optionally change passwords for security
   ```

4. **Start all services**:
   ```bash
   docker-compose up -d
   ```

5. **Check service health**:
   ```bash
   docker-compose ps
   ```

6. **View logs**:
   ```bash
   docker-compose logs -f
   ```

## Services and Ports

Once running, the following services will be available:

| Service | Port | URL | Public |
|---------|------|-----|--------|
| Gateway | 8000 | http://localhost:8000 | Yes (entry point) |
| Svedprint | 8001 | http://localhost:8001 | No (internal) |
| Svedprint Admin | 8002 | http://localhost:8002 | No (internal) |
| Svedprint Print | 8003 | http://localhost:8003 | No (internal) |
| Keycloak | 8080 | http://localhost:8080 | Yes (auth) |
| PostgreSQL | 5432 | localhost:5432 | Database |
| Redis | 6379 | localhost:6379 | Cache |

## Architecture Flow

```
User Request
     │
     ▼
┌─────────────────┐
│   Gateway :8000 │  ← Entry point (validates JWT)
└────────┬────────┘
         │
         ├──────────────────┬───────────────────┐
         ▼                  ▼                   ▼
┌──────────────┐  ┌──────────────────┐  ┌──────────────────┐
│ Svedprint    │  │ Svedprint Admin  │  │ Svedprint Print  │
│ :8001        │  │ :8002            │  │ :8003            │
└──────┬───────┘  └─────────┬────────┘  └─────────┬────────┘
       │                    │                      │
       ├────────────────────┴──────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│  Data Layer                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │Postgres  │  │Postgres  │  │Postgres  │  │
│  │Main      │  │Admin     │  │Gateway   │  │
│  │:5432     │  │:5432     │  │:5432     │  │
│  └──────────┘  └──────────┘  └──────────┘  │
│                                              │
│  ┌──────────┐                               │
│  │ Redis    │                               │
│  │ :6379    │                               │
│  └──────────┘                               │
└──────────────────────────────────────────────┘

┌──────────────┐
│  Keycloak    │  ← Authentication
│  :8080       │
└──────────────┘
```

## Databases Created

The `init-db.sh` script automatically creates these databases:

1. **svedprint_db** - Main application data (students, schools, teachers, etc.)
2. **admin_db** - Admin service data (audit logs, migrations, config)
3. **gateway_db** - Gateway request logs
4. **keycloak_db** - Keycloak authentication data

## Initial Setup Tasks

### 1. Configure Keycloak

After starting the services, configure Keycloak:

1. **Access Keycloak Admin Console**:
   - URL: http://localhost:8080
   - Username: `admin`
   - Password: `admin` (from .env)

2. **Create a Realm**:
   - Click "Create Realm"
   - Name: `svedprint`
   - Save

3. **Create a Client**:
   - Go to "Clients" → "Create Client"
   - Client ID: `svedprint-backend`
   - Client Protocol: `openid-connect`
   - Enable "Client Authentication"
   - Save

4. **Get Client Secret**:
   - Go to "Credentials" tab
   - Copy the "Client Secret"
   - Update `.env` file with `KEYCLOAK_CLIENT_SECRET=<copied-secret>`
   - Restart gateway: `docker-compose restart gateway`

5. **Create Test User**:
   - Go to "Users" → "Add User"
   - Username: `testuser`
   - Save
   - Go to "Credentials" tab
   - Set password (uncheck "Temporary")

### 2. Test the System

**Get an access token**:
```bash
curl -X POST http://localhost:8080/realms/svedprint/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=svedprint-backend" \
  -d "client_secret=<your-client-secret>" \
  -d "username=testuser" \
  -d "password=<test-password>" \
  -d "grant_type=password"
```

**Use the token to call the API**:
```bash
TOKEN="<access-token-from-above>"

curl -X GET http://localhost:8000/api/schools \
  -H "Authorization: Bearer $TOKEN"
```

## Docker Compose Commands

### Start Services
```bash
# Start all services in detached mode
docker-compose up -d

# Start specific service
docker-compose up -d svedprint

# Start with build (after code changes)
docker-compose up -d --build
```

### Stop Services
```bash
# Stop all services
docker-compose stop

# Stop specific service
docker-compose stop svedprint

# Stop and remove containers
docker-compose down

# Stop and remove containers + volumes (deletes all data)
docker-compose down -v
```

### View Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f gateway
docker-compose logs -f svedprint

# Last 100 lines
docker-compose logs --tail=100 gateway
```

### Rebuild Services
```bash
# Rebuild all services
docker-compose build

# Rebuild specific service
docker-compose build gateway

# Rebuild and restart
docker-compose up -d --build gateway
```

### Service Management
```bash
# Restart a service
docker-compose restart gateway

# Check service status
docker-compose ps

# Execute command in running container
docker-compose exec svedprint sh

# View service resource usage
docker-compose stats
```

### Database Management
```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U postgres -d svedprint_db

# Backup database
docker-compose exec postgres pg_dump -U postgres svedprint_db > backup.sql

# Restore database
docker-compose exec -T postgres psql -U postgres -d svedprint_db < backup.sql

# Connect to Redis
docker-compose exec redis redis-cli
```

## Development Workflow

### Making Code Changes

1. **Edit code in your local editor**

2. **Rebuild and restart the service**:
   ```bash
   docker-compose up -d --build svedprint
   ```

3. **View logs for debugging**:
   ```bash
   docker-compose logs -f svedprint
   ```

### Database Migrations

Migrations run automatically on service startup. To add a new migration:

1. **Create migration files**:
   ```bash
   # For svedprint service
   # Create: db/svedprint/migrations/YYYYMMDDHHMMSS_description.up.sql
   # Create: db/svedprint/migrations/YYYYMMDDHHMMSS_description.down.sql
   ```

2. **Restart the service** (migrations run on startup):
   ```bash
   docker-compose restart svedprint
   ```

### Regenerate sqlc Code

After modifying SQL queries:

1. **Install sqlc locally** (or run in container):
   ```bash
   go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
   ```

2. **Generate code**:
   ```bash
   # For svedprint
   sqlc generate -f svedprint-sqlc.yaml

   # For admin
   sqlc generate -f svedprint-admin-sqlc.yaml

   # For gateway
   sqlc generate -f gateway-sqlc.yaml
   ```

3. **Rebuild the service**:
   ```bash
   docker-compose up -d --build svedprint
   ```

## Troubleshooting

### Services won't start

**Check logs**:
```bash
docker-compose logs
```

**Common issues**:
- Port already in use: Change port in `.env`
- Database connection failed: Ensure Postgres is healthy
- Missing environment variables: Check `.env` file

### Database connection errors

**Check Postgres is running**:
```bash
docker-compose ps postgres
```

**Test connection**:
```bash
docker-compose exec postgres psql -U postgres -d svedprint_db -c "SELECT 1"
```

**Reset database**:
```bash
docker-compose down -v
docker-compose up -d postgres
```

### Keycloak not accessible

**Check Keycloak logs**:
```bash
docker-compose logs keycloak
```

**Wait for Keycloak to start** (takes 30-60 seconds):
```bash
docker-compose ps keycloak
# Wait until "healthy"
```

### Gateway returns 401 Unauthorized

**Check JWT configuration**:
- Verify `KEYCLOAK_JWKS_URL` in `.env`
- Ensure Keycloak is running
- Verify token is not expired

**Test Keycloak JWKS endpoint**:
```bash
curl http://localhost:8080/realms/svedprint/protocol/openid-connect/certs
```

### Redis connection issues

**Test Redis**:
```bash
docker-compose exec redis redis-cli ping
# Should return: PONG
```

### Service can't reach another service

**Check network**:
```bash
docker network ls
docker network inspect svedprint-go_svedprint-network
```

**Use service names** (not localhost) in environment variables:
- ✅ `http://svedprint:8001`
- ❌ `http://localhost:8001`

## Clean Slate Reset

To completely reset the environment:

```bash
# Stop and remove everything
docker-compose down -v

# Remove images
docker-compose down --rmi all

# Start fresh
docker-compose up -d
```

## Performance Tips

1. **Limit logs**: Add to `docker-compose.yml`:
   ```yaml
   logging:
     driver: "json-file"
     options:
       max-size: "10m"
       max-file: "3"
   ```

2. **Reduce build time**: Use build cache
   ```bash
   docker-compose build --parallel
   ```

3. **Monitor resources**:
   ```bash
   docker-compose stats
   ```

## Environment Variables Reference

See `.env.example` for a complete list of configuration options.

### Critical Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `KEYCLOAK_CLIENT_SECRET` | Keycloak client secret | (must set) |
| `POSTGRES_PASSWORD` | Postgres password | `password` |
| `GATEWAY_PORT` | Gateway port | `8000` |
| `GIN_MODE` | Gin mode (debug/release) | `debug` |

## Next Steps

- [Railway Deployment Guide](./RAILWAY_DEPLOYMENT.md)
- [API Documentation](./API_DOCUMENTATION.md) (if exists)
- [Development Guidelines](./GUIDELINES.md)
