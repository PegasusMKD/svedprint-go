package jwt

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWKSResponse represents the JWKS endpoint response
type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// KeycloakClaims represents the JWT claims from Keycloak
type KeycloakClaims struct {
	jwt.RegisteredClaims
	Email         string                 `json:"email"`
	EmailVerified bool                   `json:"email_verified"`
	PreferredUsername string             `json:"preferred_username"`
	GivenName     string                 `json:"given_name"`
	FamilyName    string                 `json:"family_name"`
	RealmAccess   map[string]interface{} `json:"realm_access"`
	ResourceAccess map[string]interface{} `json:"resource_access"`
}

// Validator handles JWT validation
type Validator struct {
	jwksURL    string
	realm      string
	clientID   string
	keys       map[string]*rsa.PublicKey
	mu         sync.RWMutex
	lastFetch  time.Time
	httpClient *http.Client
}

// NewValidator creates a new JWT validator
func NewValidator(jwksURL, realm, clientID string) *Validator {
	return &Validator{
		jwksURL:  jwksURL,
		realm:    realm,
		clientID: clientID,
		keys:     make(map[string]*rsa.PublicKey),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ValidateToken validates a JWT token and returns the claims
func (v *Validator) ValidateToken(ctx context.Context, tokenString string) (*KeycloakClaims, error) {
	// Refresh keys if needed (cache for 1 hour)
	if time.Since(v.lastFetch) > 1*time.Hour {
		if err := v.refreshKeys(ctx); err != nil {
			return nil, fmt.Errorf("failed to refresh JWKS keys: %w", err)
		}
	}

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenString, &KeycloakClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the key ID from token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("kid not found in token header")
		}

		// Get the public key
		v.mu.RLock()
		key, exists := v.keys[kid]
		v.mu.RUnlock()

		if !exists {
			// Try refreshing keys
			if err := v.refreshKeys(ctx); err != nil {
				return nil, fmt.Errorf("key not found and refresh failed: %w", err)
			}

			v.mu.RLock()
			key, exists = v.keys[kid]
			v.mu.RUnlock()

			if !exists {
				return nil, fmt.Errorf("key with kid %s not found", kid)
			}
		}

		return key, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	claims, ok := token.Claims.(*KeycloakClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Validate issuer
	expectedIssuer := fmt.Sprintf("%s/realms/%s", v.jwksURL[:len(v.jwksURL)-len("/protocol/openid-connect/certs")], v.realm)
	if claims.Issuer != expectedIssuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", expectedIssuer, claims.Issuer)
	}

	// Validate audience (optional, depending on your Keycloak setup)
	// Uncomment if you want strict audience validation
	// if !claims.VerifyAudience(v.clientID, true) {
	// 	return nil, errors.New("invalid audience")
	// }

	return claims, nil
}

// refreshKeys fetches and caches the public keys from Keycloak
func (v *Validator) refreshKeys(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("JWKS endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	var jwks JWKSResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS response: %w", err)
	}

	// Convert JWKs to RSA public keys
	newKeys := make(map[string]*rsa.PublicKey)
	for _, jwk := range jwks.Keys {
		if jwk.Kty != "RSA" {
			continue
		}

		key, err := jwkToRSAPublicKey(jwk)
		if err != nil {
			return fmt.Errorf("failed to convert JWK to RSA public key: %w", err)
		}

		newKeys[jwk.Kid] = key
	}

	v.mu.Lock()
	v.keys = newKeys
	v.lastFetch = time.Now()
	v.mu.Unlock()

	return nil
}

// jwkToRSAPublicKey converts a JWK to an RSA public key
func jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Decode the modulus
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode the exponent
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert to big integers
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}

// GetUserID extracts the user ID from claims
func (c *KeycloakClaims) GetUserID() string {
	return c.Subject
}

// GetUsername extracts the username from claims
func (c *KeycloakClaims) GetUsername() string {
	return c.PreferredUsername
}

// GetEmail extracts the email from claims
func (c *KeycloakClaims) GetEmail() string {
	return c.Email
}

// HasRealmRole checks if the user has a specific realm role
func (c *KeycloakClaims) HasRealmRole(role string) bool {
	if c.RealmAccess == nil {
		return false
	}

	roles, ok := c.RealmAccess["roles"].([]interface{})
	if !ok {
		return false
	}

	for _, r := range roles {
		if roleStr, ok := r.(string); ok && roleStr == role {
			return true
		}
	}

	return false
}
