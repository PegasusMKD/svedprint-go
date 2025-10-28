# Keycloak Integration Guide

This document explains how to integrate Keycloak with the svedprint-go project for authentication and authorization.

## Project Requirements

- **Client Application**: Windows desktop application (not web-based)
- **No Migration**: Starting fresh with no existing user data
- **Role Management**: All roles managed exclusively in Keycloak (no database role storage)
- **SSO Layer**: Keycloak handles authentication, registration, and user identity

## Keycloak Integration Overview

Keycloak serves as the SSO layer and handles:
- **Authentication**: Teacher login and registration
- **User Identity**: Username, email, and basic profile information
- **Authorization**: Role-based access control via Keycloak roles:
  - `teacher` - Basic teacher access
  - `print_allowed` - Permission to generate diplomas/testimonies
  - `admin` - Administrative access
- **Token Management**: JWT access tokens for API authentication

## Architecture Integration

Following your layered architecture, you'd add:

```
Windows App → Keycloak (OAuth2/OIDC) → Go API (JWT Validation) → Handler → Service → Repository
```

**Key Pattern**:
1. Windows app authenticates user with Keycloak (OAuth2 flow)
2. Windows app receives JWT access token
3. Windows app sends token with each API request
4. Go API middleware validates JWT and extracts roles
5. Handlers enforce role-based permissions

## Implementation Approach

### 1. Keycloak Setup

```bash
# Run Keycloak via Docker
docker run -p 8080:8080 \
  -e KEYCLOAK_ADMIN=admin \
  -e KEYCLOAK_ADMIN_PASSWORD=admin \
  quay.io/keycloak/keycloak:latest start-dev
```

**Keycloak Configuration**:

1. **Create Realm**: `svedprint`

2. **Create Clients**:
   - `svedprint-desktop` (public client for Windows app)
     - Access Type: `public`
     - Standard Flow Enabled: `ON`
     - Valid Redirect URIs: `http://localhost:*/callback` (or custom protocol like `svedprint://callback`)
     - Web Origins: `+` (to allow CORS)
   - `svedprint-api` (optional, for service accounts if needed)
     - Access Type: `confidential`

3. **Define Realm Roles**:
   - `teacher` - Basic teacher access
   - `admin` - Administrative access
   - `print_allowed` - Permission to generate diplomas

4. **Enable User Registration**:
   - Realm Settings → Login → User registration: `ON`
   - Configure registration flow and required user attributes

5. **Configure Token Settings**:
   - Realm Settings → Tokens → Access Token Lifespan: `15 minutes` (adjust as needed)
   - Refresh Token settings for desktop app

### 2. Go Dependencies

```bash
go get github.com/golang-jwt/jwt/v5
go get github.com/coreos/go-oidc/v3/oidc
go get golang.org/x/oauth2
```

### 3. Windows Desktop Authentication Flow

The Windows application will use **OAuth2 Authorization Code Flow with PKCE**:

```
1. User clicks "Login" in Windows app
2. App opens browser to Keycloak login page (with PKCE code_challenge)
3. User authenticates and grants permissions
4. Keycloak redirects to http://localhost:<port>/callback with authorization code
5. Windows app exchanges code for access token (with PKCE code_verifier)
6. App stores tokens securely (Windows Credential Manager)
7. App includes access token in Authorization header for all API calls
```

**Windows App Libraries** (if using .NET):
- `IdentityModel.OidcClient` - OAuth2/OIDC client for desktop apps
- Alternative: `Microsoft.Identity.Client` (MSAL)

### 4. Code Structure

```
internal/
├── auth/
│   ├── middleware.go      # JWT validation middleware
│   ├── keycloak.go        # Keycloak JWT verifier setup
│   └── context.go         # Extract user info from JWT
├── teacher/
│   └── service.go         # Link Keycloak users to school business data
```

### 5. Middleware Example

```go
// internal/auth/middleware.go
package auth

import (
    "context"
    "github.com/gin-gonic/gin"
    "github.com/coreos/go-oidc/v3/oidc"
    "strings"
)

type AuthMiddleware struct {
    verifier *oidc.IDTokenVerifier
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "missing authorization header"})
            c.Abort()
            return
        }

        token := strings.TrimPrefix(authHeader, "Bearer ")

        // Verify JWT with Keycloak
        idToken, err := m.verifier.Verify(c.Request.Context(), token)
        if err != nil {
            c.JSON(401, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }

        // Extract claims
        var claims struct {
            Sub           string `json:"sub"`
            Email         string `json:"email"`
            PreferredName string `json:"preferred_username"`
            RealmAccess   struct {
                Roles []string `json:"roles"`
            } `json:"realm_access"`
        }

        if err := idToken.Claims(&claims); err != nil {
            c.JSON(500, gin.H{"error": "failed to parse claims"})
            c.Abort()
            return
        }

        // Store in context for handlers
        c.Set("user_id", claims.Sub)           // Keycloak user UUID
        c.Set("user_email", claims.Email)
        c.Set("username", claims.PreferredName)
        c.Set("user_roles", claims.RealmAccess.Roles)

        c.Next()
    }
}

func (m *AuthMiddleware) RequireRole(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        roles, _ := c.Get("user_roles")
        userRoles := roles.([]string)

        for _, r := range userRoles {
            if r == role {
                c.Next()
                return
            }
        }

        c.JSON(403, gin.H{"error": "insufficient permissions"})
        c.Abort()
    }
}
```

### 6. Router Integration (cmd/api.go)

```go
router := gin.Default()

// Initialize auth middleware
authMiddleware := auth.NewAuthMiddleware(keycloakURL, keycloakRealm)

// Health check (no auth required)
router.GET("/health", healthHandler.Check)

// All API endpoints require authentication
api := router.Group("/api")
api.Use(authMiddleware.RequireAuth())
{
    // Endpoints that require 'teacher' role
    teachers := api.Group("")
    teachers.Use(authMiddleware.RequireRole("teacher"))
    {
        teachers.GET("/students", studentHandler.List)
        teachers.GET("/students/:id", studentHandler.GetByID)
        teachers.GET("/schools", schoolHandler.List)
    }

    // Endpoints that require 'print_allowed' role
    printing := api.Group("/print")
    printing.Use(authMiddleware.RequireRole("print_allowed"))
    {
        printing.POST("/diploma/:student_id", diplomaHandler.Generate)
        printing.POST("/testimony/:student_id", testimonyHandler.Generate)
    }

    // Endpoints that require 'admin' role
    admin := api.Group("/admin")
    admin.Use(authMiddleware.RequireRole("admin"))
    {
        admin.POST("/students", studentHandler.Create)
        admin.PUT("/students/:id", studentHandler.Update)
        admin.DELETE("/students/:id", studentHandler.Delete)
        admin.POST("/schools", schoolHandler.Create)
    }
}
```

### 7. Database Schema Changes

Since roles are managed in Keycloak, simplify the `teacher` table:

```sql
-- Remove password column (authentication is via Keycloak)
ALTER TABLE teacher DROP COLUMN password;
ALTER TABLE teacher DROP COLUMN print_allowed;  -- Roles managed in Keycloak

-- Ensure keycloak_sub exists and is the primary identifier
-- (Already exists as teacher_uuid in your schema, so map it to Keycloak's 'sub' claim)
ALTER TABLE teacher ADD CONSTRAINT teacher_keycloak_sub_unique UNIQUE (teacher_uuid);

-- Add username from Keycloak for display purposes
ALTER TABLE teacher ADD COLUMN username TEXT NOT NULL;
```

**Teacher Table Purpose**: Store business-specific data only
- `teacher_uuid` → Maps to Keycloak `sub` (user ID)
- `username` → From Keycloak for display
- `email` → From Keycloak
- `school_uuid` → Business relationship (which school they belong to)
- Other business fields as needed

**Role Checking**: All role checks happen in the Go API middleware by reading JWT claims - no database queries needed.

### 8. Teacher Service Logic

When a teacher first accesses the system (after Keycloak authentication):

```go
// internal/teacher/service.go
func (s *TeacherService) EnsureTeacherExists(ctx context.Context, keycloakSub, username, email string) error {
    // Check if teacher exists in database
    teacher, err := s.repo.GetByKeycloakSub(ctx, keycloakSub)
    if err == sql.ErrNoRows {
        // First-time login: create teacher record
        return s.repo.Create(ctx, &Teacher{
            UUID:     keycloakSub,  // Use Keycloak sub as UUID
            Username: username,
            Email:    email,
            // school_uuid can be assigned by admin later
        })
    }
    return err
}
```

Call this in a middleware or during first API request to ensure teacher record exists.

### 9. Configuration (environment variables)

**Go API** (.env or config):
```bash
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=svedprint
```

**Windows Desktop App** (config file or settings):
```bash
KEYCLOAK_URL=http://localhost:8080
KEYCLOAK_REALM=svedprint
KEYCLOAK_CLIENT_ID=svedprint-desktop
API_BASE_URL=http://localhost:8000/api
```

## Windows Desktop App: Authentication Flow Example

**Using .NET and IdentityModel.OidcClient**:

```csharp
using IdentityModel.OidcClient;

public class AuthService
{
    private readonly OidcClient _client;

    public AuthService()
    {
        _client = new OidcClient(new OidcClientOptions
        {
            Authority = "http://localhost:8080/realms/svedprint",
            ClientId = "svedprint-desktop",
            RedirectUri = "http://localhost:5000/callback",
            Scope = "openid profile email",

            // Use system browser for login
            Browser = new SystemBrowser()
        });
    }

    public async Task<LoginResult> LoginAsync()
    {
        var result = await _client.LoginAsync(new LoginRequest());

        if (result.IsError)
        {
            throw new Exception(result.Error);
        }

        // Store tokens securely
        SaveTokens(result.AccessToken, result.RefreshToken);

        return result;
    }

    public string GetAccessToken()
    {
        // Retrieve from secure storage
        return LoadAccessToken();
    }
}

// Making API calls with token
var httpClient = new HttpClient();
httpClient.DefaultRequestHeaders.Authorization =
    new AuthenticationHeaderValue("Bearer", authService.GetAccessToken());

var response = await httpClient.GetAsync("http://localhost:8000/api/students");
```

## Key Design Decisions

✅ **Decided**:
1. **User Storage**: Keycloak is the source of truth for authentication and roles
2. **Role Management**: All roles managed exclusively in Keycloak (no `print_allowed` flag in DB)
3. **School Association**: Store teacher-school relationship in PostgreSQL (business logic)
4. **Token Storage**: Windows Credential Manager or encrypted local storage in desktop app
5. **No Migration**: Fresh start with no existing user data to migrate

## Benefits for Your Use Case

- **Separation of Concerns**: Authentication logic separate from business logic
- **Self-Service**: Teachers can register and reset passwords via Keycloak UI
- **Centralized Role Management**: Admins manage roles in one place (Keycloak)
- **Desktop App Support**: Native OAuth2 flows for Windows applications
- **SSO Ready**: Can integrate with school district LDAP/AD later
- **Audit Trail**: Keycloak logs all authentication events
- **No Password Management**: Your Go API never handles passwords

## Implementation Checklist

### Phase 1: Keycloak Setup
- [ ] Run Keycloak via Docker
- [ ] Create `svedprint` realm
- [ ] Create `svedprint-desktop` public client
- [ ] Define roles: `teacher`, `admin`, `print_allowed`
- [ ] Enable user registration
- [ ] Test login flow manually

### Phase 2: Go API Integration
- [ ] Install Go dependencies (`go-oidc`, `pgx`)
- [ ] Create `internal/auth/` package
- [ ] Implement JWT validation middleware
- [ ] Implement role-checking middleware
- [ ] Update `cmd/api.go` to use auth middleware
- [ ] Update teacher table schema (remove password/print_allowed)
- [ ] Test API with Postman using JWT tokens

### Phase 3: Windows Desktop App
- [ ] Set up .NET project with `IdentityModel.OidcClient`
- [ ] Implement login/logout flow
- [ ] Implement secure token storage
- [ ] Implement token refresh logic
- [ ] Add Authorization header to all API calls
- [ ] Handle 401/403 responses (re-login prompt)

### Phase 4: Production Readiness
- [ ] Set up docker-compose with Keycloak + PostgreSQL
- [ ] Configure Keycloak with production database (not H2)
- [ ] Set up HTTPS for Keycloak
- [ ] Configure proper token lifetimes
- [ ] Set up Keycloak backups
- [ ] Document user registration and role assignment process

## Next Steps

Ready to implement? I can help you with:

1. **Create docker-compose.yml** with Keycloak + PostgreSQL setup
2. **Implement Go middleware** for JWT validation and role checking
3. **Create database migration** to update teacher table schema
4. **Write .NET authentication example** for the Windows desktop app
5. **Set up initial Keycloak configuration** (realm export JSON)

Let me know which part you'd like to start with!
