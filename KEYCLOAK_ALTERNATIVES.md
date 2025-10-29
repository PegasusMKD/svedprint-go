# Keycloak Alternatives for Svedprint

## Overview

This document evaluates authentication solutions for the Svedprint student management and diploma printing system. The system requires teacher authentication with role-based access control (teachers, admins) and needs to be cost-effective for deployment on Railway.

---

## Current Situation

**Keycloak Issues:**
- High memory usage (~400-800MB even when idle)
- Complex setup and configuration
- Overkill for current use case
- Contributing ~$2-4/month to Railway costs when idle

**Requirements:**
- Teacher/Admin authentication
- Role-based access control (RBAC)
- JWT token support
- Integration with Go backend
- Cost-effective for small-medium scale deployment
- Support for ~50-500 users (teachers across multiple schools)

---

## Options

### 1. Clerk

**Website:** https://clerk.com

**Description:** Modern, developer-friendly authentication platform with beautiful pre-built UI components.

#### Pros
- **Excellent Developer Experience** - Best-in-class documentation and Go SDK
- **Generous Free Tier** - 10,000 monthly active users (MAU) free
- **Pre-built Components** - Beautiful login/signup UI out of the box
- **Organizations/Multi-tenancy** - Built-in support (perfect for schools)
- **No Infrastructure** - Fully managed, zero maintenance
- **Fast Integration** - Can be set up in under 30 minutes
- **Modern Features** - MFA, social login, magic links included
- **Great Go Support** - Official clerk-sdk-go package

#### Cons
- **Vendor Lock-in** - Proprietary platform, hard to migrate away
- **Pricing After Free Tier** - $25/month for 1,000 MAU (can add up)
- **Limited Customization** - UI customization has some constraints
- **US-Centric** - Some features prioritize US market

#### Cost
- **Free:** 10,000 MAU
- **Pro:** $25/month for 1,000 MAU, then $0.02/MAU
- **For your use case:** FREE (unless you exceed 10k teachers)

#### Integration Complexity: ⭐⭐⭐⭐⭐ (Very Easy)

---

### 2. Supabase Auth

**Website:** https://supabase.com

**Description:** Open-source Firebase alternative with built-in authentication powered by GoTrue.

#### Pros
- **Completely Free (Self-hosted)** - Open source, can run on Railway
- **Generous Managed Free Tier** - 50,000 MAU free
- **PostgreSQL Native** - You're already using PostgreSQL
- **Row Level Security (RLS)** - Database-level permissions
- **Open Source** - Can self-host, no vendor lock-in
- **Full Stack Solution** - Also provides database, storage, edge functions
- **Good Go Support** - Community libraries available (supabase-go)
- **Modern Auth Flows** - Magic links, social auth, phone auth

#### Cons
- **Less Polish than Clerk** - UI components not as refined
- **Learning Curve** - RLS policies can be complex
- **Go SDK Not Official** - Community-maintained libraries
- **Additional Infrastructure** - Self-hosting adds complexity
- **Limited RBAC** - Basic roles, not as advanced as Keycloak

#### Cost
- **Free (Managed):** 50,000 MAU, 500MB database, 1GB storage
- **Pro:** $25/month for 100,000 MAU
- **Self-hosted:** Cost of hosting (Railway ~$1-2/month for auth service)
- **For your use case:** FREE

#### Integration Complexity: ⭐⭐⭐⭐ (Easy)

---

### 3. Auth0 (by Okta)

**Website:** https://auth0.com

**Description:** Industry-standard enterprise authentication platform, battle-tested and feature-rich.

#### Pros
- **Industry Standard** - Used by major enterprises (reliability)
- **Generous Free Tier** - 7,500 MAU free
- **Comprehensive Features** - Every auth feature you could need
- **Excellent Documentation** - Mature ecosystem
- **Strong RBAC** - Sophisticated role and permission system
- **Great Go Support** - Official go-auth0 SDK
- **Compliance** - SOC2, HIPAA, GDPR ready
- **Universal Login** - Customizable hosted login pages

#### Cons
- **Complex UI** - Dashboard can be overwhelming
- **Slower Innovation** - Since Okta acquisition, less cutting-edge
- **Expensive at Scale** - $240/month after free tier (B2B)
- **Overkill for Small Apps** - Many features you won't use
- **Setup Complexity** - More configuration needed than Clerk

#### Cost
- **Free:** 7,500 MAU, unlimited logins
- **Essentials:** $35/month for 500 MAU (B2C) or $240/month (B2B)
- **For your use case:** FREE (unless you exceed 7,500 teachers)

#### Integration Complexity: ⭐⭐⭐ (Moderate)

---

### 4. Custom JWT Authentication (DIY)

**Description:** Build your own authentication using JWT tokens, bcrypt password hashing, and PostgreSQL.

#### Pros
- **Zero External Costs** - No third-party services
- **Full Control** - Complete customization freedom
- **No Vendor Lock-in** - Your code, your rules
- **Simple Stack** - Just Go + PostgreSQL
- **Privacy** - All user data stays in your database
- **Learning Opportunity** - Understand auth deeply
- **Perfect Fit** - Exactly what you need, nothing more

#### Cons
- **Development Time** - 2-4 weeks to build properly
- **Security Responsibility** - You own all security vulnerabilities
- **Maintenance Burden** - Password resets, MFA, etc. all on you
- **No Pre-built UI** - Must build login forms yourself
- **Missing Features** - No social login, magic links, etc. (unless you build them)
- **Compliance Risk** - Harder to prove security standards

#### Cost
- **Development:** 2-4 weeks of your time
- **Ongoing:** $0 (already using PostgreSQL)
- **For your use case:** FREE

#### Integration Complexity: ⭐⭐ (Complex - requires building)

**Implementation Highlights:**
- Use `golang-jwt/jwt/v5` for JWT tokens
- Use `golang.org/x/crypto/bcrypt` for password hashing
- Store sessions in Redis (you already have it)
- Implement refresh tokens for security
- Add rate limiting for login attempts

---

### 5. SuperTokens

**Website:** https://supertokens.com

**Description:** Open-source authentication with managed and self-hosted options, designed for developers.

#### Pros
- **Open Source** - MIT license, can self-host completely free
- **Generous Free Tier (Managed)** - 5,000 MAU free
- **Session Management** - Best-in-class session handling
- **Pre-built UI** - React/Vue/Angular components (can adapt for Go templates)
- **Good Go Support** - Official supertokens-golang SDK
- **Security First** - Built to prevent common auth vulnerabilities
- **Self-host Option** - Run on Railway alongside your app
- **Modern Architecture** - Designed for microservices

#### Cons
- **Smaller Ecosystem** - Less mature than Auth0/Clerk
- **Limited Social Auth** - Fewer providers than competitors
- **Documentation Gaps** - Some advanced features poorly documented
- **Community Size** - Smaller community for troubleshooting
- **UI Not Great** - Pre-built components need heavy styling

#### Cost
- **Free (Managed):** 5,000 MAU
- **Growth:** $30/month for 10,000 MAU
- **Self-hosted:** Cost of hosting (~$1-2/month on Railway)
- **For your use case:** FREE

#### Integration Complexity: ⭐⭐⭐⭐ (Easy)

---

### 6. Ory (Kratos + Hydra)

**Website:** https://www.ory.sh

**Description:** Open-source identity and access management, built for cloud-native applications. Direct Keycloak alternative.

#### Pros
- **Open Source** - Apache 2.0 license, fully auditable
- **Cloud Native** - Built for Kubernetes/microservices
- **Security Focused** - Built by security experts
- **No Vendor Lock-in** - Self-host forever for free
- **Headless Architecture** - API-first, maximum flexibility
- **Good Go Support** - Written in Go, excellent Go SDK
- **Advanced Features** - MFA, account recovery, identity management
- **OAuth2/OIDC** - Full OAuth2 and OpenID Connect support

#### Cons
- **Steep Learning Curve** - Most complex to set up
- **No Pre-built UI** - Completely headless, must build all UI
- **Configuration Heavy** - Lots of YAML configuration needed
- **Documentation Overwhelming** - Very technical, less beginner-friendly
- **Resource Usage** - Still heavier than managed solutions (~200-400MB)
- **Multiple Components** - Need both Kratos (identity) and Hydra (OAuth2)

#### Cost
- **Self-hosted:** Free + hosting costs (~$2-4/month on Railway)
- **Ory Cloud:** $30/month for 1,000 MAU
- **For your use case:** FREE (self-hosted) or $30/month (managed)

#### Integration Complexity: ⭐ (Very Complex)

---

### 7. FusionAuth

**Website:** https://fusionauth.io

**Description:** Developer-focused identity and access management platform with both managed and self-hosted options.

#### Pros
- **Generous Free Tier** - Unlimited users on Community Edition
- **Self-host Option** - Docker image available, easy to deploy
- **Feature Rich** - Comparable to Auth0/Keycloak
- **Good Documentation** - Clear guides and examples
- **Go SDK Available** - Community-maintained go-fusionauth
- **Advanced RBAC** - Sophisticated permission system
- **Low Resource Usage** - ~300-400MB RAM (lighter than Keycloak)
- **Migration Tools** - Can migrate from other auth systems

#### Cons
- **Self-hosting Required for Free** - Managed version is expensive
- **Still Heavy** - Not as lightweight as managed solutions
- **Slower Startup** - Cold start ~15-20 seconds
- **Less Modern UI** - Admin dashboard looks dated
- **Smaller Community** - Less popular than Auth0/Clerk

#### Cost
- **Community (Self-hosted):** FREE
- **Hosted Starter:** $75/month
- **For your use case:** FREE (self-hosted) ~$2-3/month hosting on Railway

#### Integration Complexity: ⭐⭐⭐ (Moderate)

---

## Comparison Matrix

| Solution | Free Tier | Monthly Cost (Idle) | Setup Time | Go Support | RBAC | Best For |
|----------|-----------|---------------------|------------|------------|------|----------|
| **Clerk** | 10k MAU | $0 | 30 min | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | Quick launch, modern UX |
| **Supabase** | 50k MAU | $0 | 2 hours | ⭐⭐⭐⭐ | ⭐⭐⭐ | Full-stack apps, PostgreSQL fans |
| **Auth0** | 7.5k MAU | $0 | 1-2 hours | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Enterprise features, reliability |
| **Custom JWT** | Unlimited | $0 | 2-4 weeks | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Full control, learning |
| **SuperTokens** | 5k MAU | $0-1 | 1-2 hours | ⭐⭐⭐⭐ | ⭐⭐⭐ | Balance of control/ease |
| **Ory** | Unlimited | $2-4 | 1-2 days | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Security-critical apps |
| **FusionAuth** | Unlimited | $2-3 | 2-4 hours | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Keycloak replacement |
| **Keycloak** (current) | Unlimited | $3-5 | 1-2 days | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Enterprise SSO, complex RBAC |

---

## My Recommendation: **Clerk**

### Why Clerk is Best for Svedprint

#### 1. **Perfect for Your Scale**
- You have schools with teachers (likely < 1,000 total users)
- 10,000 MAU free tier means you'll never pay unless you massively scale
- Current Keycloak setup is overkill for your user count

#### 2. **Fastest Time to Value**
- Can be integrated in under 1 hour
- Pre-built components mean no UI work needed
- Excellent Go SDK with clear examples

#### 3. **Built-in Multi-tenancy (Organizations)**
- Perfect for your school-based model
- Each school can be an "organization"
- Teachers belong to their school's organization
- Admins can manage multiple schools

#### 4. **Zero Infrastructure Costs**
- Removes $3-5/month Keycloak costs on Railway
- No additional services to manage
- No memory/CPU usage on your infrastructure

#### 5. **Better Developer Experience**
- You can focus on building diploma features, not auth
- Excellent documentation with Go examples
- Active community and fast support

#### 6. **Production Ready Features**
- MFA out of the box (if needed later)
- Email verification included
- Password reset flows handled
- Session management optimized

### Implementation Plan with Clerk

```go
// 1. Install Clerk SDK
// go get github.com/clerk/clerk-sdk-go/v2

// 2. Middleware for Gateway
func ClerkAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        sessionToken := c.GetHeader("Authorization")

        // Verify with Clerk
        client, _ := clerk.NewClient(os.Getenv("CLERK_SECRET_KEY"))
        session, err := client.Sessions().Verify(sessionToken)

        if err != nil {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }

        // Set user info in context
        c.Set("user_id", session.UserID)
        c.Next()
    }
}

// 3. Role checking
func RequireRole(role string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")

        client, _ := clerk.NewClient(os.Getenv("CLERK_SECRET_KEY"))
        user, _ := client.Users().Read(userID)

        if user.PublicMetadata["role"] != role {
            c.JSON(403, gin.H{"error": "Forbidden"})
            c.Abort()
            return
        }

        c.Next()
    }
}

// 4. Usage in routes
router.GET("/admin/schools", ClerkAuthMiddleware(), RequireRole("admin"), listSchools)
router.GET("/teacher/students", ClerkAuthMiddleware(), RequireRole("teacher"), listStudents)
```

### Migration from Keycloak to Clerk

1. **Export Keycloak users** (if you have any in production)
2. **Import to Clerk** using their bulk import API
3. **Update Gateway** to use Clerk middleware
4. **Remove Keycloak** service from Railway
5. **Save $3-5/month** immediately

### When NOT to Use Clerk

Consider alternatives if:
- You need 100% self-hosted (data sovereignty requirements) → Use **Ory** or **Custom JWT**
- You anticipate >10,000 active teachers → Use **Auth0** (higher free tier)
- You want to avoid any external dependencies → Use **Custom JWT**
- You need complex enterprise SSO (SAML, LDAP) → Stick with **Keycloak**

---

## Runner-up: Supabase Auth

If you want more control and PostgreSQL integration, **Supabase Auth** is an excellent second choice:

- 50k MAU free tier (5x more than Clerk)
- Open source (can self-host if needed later)
- PostgreSQL-native (you're already using it)
- Row-level security for data permissions

**Trade-off:** Slightly more setup time, less polished UI components.

---

## Decision Matrix

**Choose Clerk if:** You want the fastest, easiest solution with great UX
**Choose Supabase if:** You want PostgreSQL integration and more control
**Choose Auth0 if:** You need enterprise features and compliance
**Choose Custom JWT if:** You want zero external dependencies and have time
**Choose Ory if:** You need maximum security and flexibility (and have patience)
**Keep Keycloak if:** You need enterprise SSO or complex federation

---

## Estimated Cost Savings

**Current (Keycloak on Railway):**
- Idle: ~$4-5/month
- Active: ~$8-12/month

**After switching to Clerk:**
- Idle: ~$0.50-1/month (just your Go services)
- Active: ~$2-4/month
- **Savings: ~$4-8/month** (50-80% reduction)

**Break-even point:** Immediate (Clerk is free for your scale)

---

## Next Steps

If you choose **Clerk** (recommended):

1. Sign up at https://clerk.com
2. Create a new application
3. Get your API keys (Secret Key and Publishable Key)
4. Install Go SDK: `go get github.com/clerk/clerk-sdk-go/v2`
5. Update Gateway middleware (example above)
6. Test locally with Clerk's development instance
7. Deploy to Railway
8. Remove Keycloak service
9. Monitor cost reduction in Railway dashboard

**Estimated implementation time: 2-4 hours**

---

## Conclusion

For the Svedprint project, **Clerk** offers the best balance of:
- ✅ Cost (free for your scale)
- ✅ Developer experience (fastest integration)
- ✅ Features (multi-tenancy perfect for schools)
- ✅ Reliability (managed service, zero maintenance)
- ✅ Scalability (room to grow to 10k users)

**Recommendation: Switch from Keycloak to Clerk and save ~$4-8/month while improving developer experience.**
