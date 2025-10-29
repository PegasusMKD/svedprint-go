# Railway Deployment Guide

This guide explains how to deploy the svedprint microservices architecture on Railway.

## Architecture Overview

The system consists of:
- **4 Go Services**: gateway, svedprint, svedprint-admin, svedprint-print
- **3 Databases**: PostgreSQL instances (can use Railway's Postgres plugin)
- **1 Redis**: For caching (Railway Redis plugin)
- **1 Keycloak**: For authentication

## Deployment Steps

### 1. Create a New Railway Project

```bash
railway login
railway init
```

### 2. Add Database Services

#### PostgreSQL (3 instances needed)

Railway doesn't support multiple databases from one Postgres instance easily, so create 3 separate Postgres plugins:

```bash
# In Railway Dashboard:
# 1. Click "New" → "Database" → "Add PostgreSQL" (for svedprint_db)
# 2. Click "New" → "Database" → "Add PostgreSQL" (for admin_db)
# 3. Click "New" → "Database" → "Add PostgreSQL" (for gateway_db)
```

**Alternative**: Use a single PostgreSQL instance and create multiple databases manually.

#### Redis

```bash
# In Railway Dashboard:
# Click "New" → "Database" → "Add Redis"
```

### 3. Deploy Keycloak

Create a new service from a Docker image:

```bash
# In Railway Dashboard:
# 1. Click "New" → "Empty Service"
# 2. Name it "keycloak"
# 3. In Settings → Deploy, set:
#    - Source: Docker Image
#    - Image: quay.io/keycloak/keycloak:23.0
#    - Start Command: start
```

Set these environment variables for Keycloak:
```
KEYCLOAK_ADMIN=admin
KEYCLOAK_ADMIN_PASSWORD=<generate-secure-password>
KC_DB=postgres
KC_DB_URL=<postgres-connection-from-railway>
KC_DB_USERNAME=<postgres-user>
KC_DB_PASSWORD=<postgres-password>
KC_HOSTNAME_STRICT=false
KC_HTTP_ENABLED=true
KC_PROXY=edge
```

Generate a domain for Keycloak (e.g., `auth.svedprint.com`).

### 4. Deploy Gateway Service

```bash
# In Railway Dashboard:
# 1. Click "New" → "GitHub Repo" (connect your repo)
# 2. Name it "gateway"
# 3. Set build arguments:
#    SERVICE_NAME=gateway
```

**Environment Variables for Gateway:**
```
PORT=8000
SERVICE_NAME=gateway
DATABASE_URL=${{Postgres-Gateway.DATABASE_URL}}
GATEWAY_DATABASE_URL=${{Postgres-Gateway.DATABASE_URL}}
KEYCLOAK_URL=${{keycloak.RAILWAY_PUBLIC_DOMAIN}}
KEYCLOAK_REALM=svedprint
KEYCLOAK_CLIENT_ID=svedprint-backend
KEYCLOAK_CLIENT_SECRET=<your-client-secret>
KEYCLOAK_JWKS_URL=${{keycloak.RAILWAY_PUBLIC_DOMAIN}}/realms/svedprint/protocol/openid-connect/certs
SVEDPRINT_SERVICE_URL=${{svedprint.RAILWAY_PRIVATE_DOMAIN}}
SVEDPRINT_ADMIN_SERVICE_URL=${{svedprint-admin.RAILWAY_PRIVATE_DOMAIN}}
SVEDPRINT_PRINT_SERVICE_URL=${{svedprint-print.RAILWAY_PRIVATE_DOMAIN}}
GIN_MODE=release
LOG_LEVEL=info
```

**Important**:
- Make the Gateway service **public** (generate a domain like `api.svedprint.com`)
- This is the ONLY public-facing service

### 5. Deploy Svedprint Main Service

```bash
# In Railway Dashboard:
# 1. Click "New" → "GitHub Repo" (same repo, different service)
# 2. Name it "svedprint"
# 3. Set build arguments:
#    SERVICE_NAME=svedprint
```

**Environment Variables for Svedprint:**
```
PORT=8001
SERVICE_NAME=svedprint
DATABASE_URL=${{Postgres-Main.DATABASE_URL}}
DATABASE_MAX_CONNS=25
DATABASE_MAX_IDLE_CONNS=10
DATABASE_CONN_MAX_LIFETIME=5m
REDIS_ADDR=${{Redis.RAILWAY_PRIVATE_DOMAIN}}:6379
REDIS_PASSWORD=${{Redis.REDIS_PASSWORD}}
REDIS_DB=0
REDIS_TTL=10m
GIN_MODE=release
LOG_LEVEL=info
```

**Important**: Keep this service **private** (use Railway's internal network).

### 6. Deploy Svedprint Admin Service

```bash
# In Railway Dashboard:
# 1. Click "New" → "GitHub Repo"
# 2. Name it "svedprint-admin"
# 3. Set build arguments:
#    SERVICE_NAME=svedprint-admin
```

**Environment Variables for Svedprint Admin:**
```
PORT=8002
SERVICE_NAME=svedprint-admin
DATABASE_URL=${{Postgres-Admin.DATABASE_URL}}
DATABASE_MAX_CONNS=25
DATABASE_MAX_IDLE_CONNS=10
DATABASE_CONN_MAX_LIFETIME=5m
SVEDPRINT_SERVICE_URL=${{svedprint.RAILWAY_PRIVATE_DOMAIN}}
GIN_MODE=release
LOG_LEVEL=info
```

**Important**: Keep this service **private**.

### 7. Deploy Svedprint Print Service

```bash
# In Railway Dashboard:
# 1. Click "New" → "GitHub Repo"
# 2. Name it "svedprint-print"
# 3. Set build arguments:
#    SERVICE_NAME=svedprint-print
```

**Environment Variables for Svedprint Print:**
```
PORT=8003
SERVICE_NAME=svedprint-print
SVEDPRINT_SERVICE_URL=${{svedprint.RAILWAY_PRIVATE_DOMAIN}}
GIN_MODE=release
LOG_LEVEL=info
```

**Important**: Keep this service **private**.

## Build Configuration

For each Go service, add a `railway.toml` file in the root (or configure in Railway dashboard):

### Gateway
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "Dockerfile"

[build.args]
SERVICE_NAME = "gateway"

[deploy]
startCommand = "/app/service"
```

### Svedprint
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "Dockerfile"

[build.args]
SERVICE_NAME = "svedprint"

[deploy]
startCommand = "/app/service"
```

### Svedprint Admin
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "Dockerfile"

[build.args]
SERVICE_NAME = "svedprint-admin"

[deploy]
startCommand = "/app/service"
```

### Svedprint Print
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "Dockerfile"

[build.args]
SERVICE_NAME = "svedprint-print"

[deploy]
startCommand = "/app/service"
```

## Railway Service Variables

Railway uses service references for internal networking:

- `${{service.RAILWAY_PRIVATE_DOMAIN}}` - Internal domain (use this for inter-service communication)
- `${{service.RAILWAY_PUBLIC_DOMAIN}}` - Public domain (only for gateway and keycloak)

## Network Configuration

```
┌─────────────────────────────────────────────┐
│  Internet                                   │
└────────┬────────────────────────────────────┘
         │
         ▼
┌─────────────────────┐
│  Gateway (Public)   │ ← api.svedprint.com
│  - Auth Validation  │
│  - Request Logging  │
└────────┬────────────┘
         │
         │ (Private Network)
         ▼
┌────────────────────────────────────────┐
│   Internal Services (Private)          │
│                                        │
│  ┌──────────────┐  ┌────────────────┐ │
│  │  Svedprint   │  │ Svedprint-Admin│ │
│  │  (Main API)  │  │  (Admin API)   │ │
│  └──────┬───────┘  └────────────────┘ │
│         │                              │
│  ┌──────▼─────────┐                   │
│  │ Svedprint-Print│                   │
│  │  (PDF Gen)     │                   │
│  └────────────────┘                   │
└────────────────────────────────────────┘
         │
         ▼
┌────────────────────────────────────────┐
│   Data Layer                           │
│                                        │
│  ┌──────────┐  ┌──────────┐          │
│  │Postgres  │  │Postgres  │          │
│  │(Main DB) │  │(Admin DB)│          │
│  └──────────┘  └──────────┘          │
│                                        │
│  ┌──────────┐  ┌──────────┐          │
│  │Postgres  │  │  Redis   │          │
│  │(Gateway) │  │ (Cache)  │          │
│  └──────────┘  └──────────┘          │
└────────────────────────────────────────┘
         │
         │
┌────────▼───────┐
│   Keycloak     │ ← auth.svedprint.com
│ (Public/Mixed) │
└────────────────┘
```

## Custom Domains

1. **Gateway**: `api.svedprint.com` - Public facing
2. **Keycloak**: `auth.svedprint.com` - Public facing (for auth)
3. **All other services**: Use Railway's internal domains (not exposed)

## Environment Variables Checklist

Before deploying, ensure you have:

- [ ] Created 3 PostgreSQL databases (or 3 separate instances)
- [ ] Created 1 Redis instance
- [ ] Set up Keycloak with proper database connection
- [ ] Generated Keycloak client credentials
- [ ] Set all environment variables for each service
- [ ] Configured service-to-service references using `${{service.RAILWAY_PRIVATE_DOMAIN}}`
- [ ] Made only Gateway and Keycloak public
- [ ] Set up custom domains for Gateway and Keycloak

## Database Migrations

Migrations run automatically on service startup via the `pkg/database/migrate.go` helper.

Each service is responsible for its own database migrations:
- **Gateway**: `db/gateway/migrations/`
- **Svedprint**: `db/svedprint/migrations/`
- **Svedprint Admin**: `db/svedprint-admin/migrations/`

## Monitoring and Logs

Access logs for each service in Railway:
```bash
railway logs --service=gateway
railway logs --service=svedprint
railway logs --service=svedprint-admin
railway logs --service=svedprint-print
```

## Scaling

Each service can be scaled independently in Railway:
- Gateway: Scale based on traffic
- Svedprint: Scale based on API load
- Svedprint Admin: Usually 1 instance is sufficient
- Svedprint Print: Scale based on PDF generation demand

## Cost Optimization

- Use Railway's sleep feature for non-production environments
- Monitor database connection pool sizes
- Adjust Redis TTL based on your caching needs
- Consider using Railway's Hobby plan for development

## Troubleshooting

### Service can't connect to database
- Check that `DATABASE_URL` is correctly set
- Verify the Postgres service is running
- Check logs for connection errors

### Gateway returns 401 errors
- Verify Keycloak is running and accessible
- Check `KEYCLOAK_JWKS_URL` is correct
- Ensure JWT tokens are valid

### Services can't communicate
- Verify you're using `RAILWAY_PRIVATE_DOMAIN` for internal calls
- Check that service references are correct (e.g., `${{svedprint.RAILWAY_PRIVATE_DOMAIN}}`)
- Ensure services are in the same Railway project

### Migrations failing
- Check that migration files are correctly copied in Dockerfile
- Verify database connection string
- Check logs for specific migration errors

## Security Checklist

- [ ] Change all default passwords
- [ ] Use Railway's secret management for sensitive values
- [ ] Enable HTTPS for public domains
- [ ] Keep internal services private (no public domain)
- [ ] Regularly update Keycloak
- [ ] Use strong Keycloak client secrets
- [ ] Enable database backups in Railway
- [ ] Set up monitoring and alerting
