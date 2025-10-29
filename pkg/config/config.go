package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServiceName string
	Port        string
	GinMode     string

	DatabaseURL          string
	DatabaseMaxConns     int
	DatabaseMaxIdleConns int
	DatabaseConnLifetime time.Duration

	RedisAddr     string
	RedisPassword string
	RedisDB       int
	RedisTTL      time.Duration

	KeycloakURL          string
	KeycloakRealm        string
	KeycloakClientID     string
	KeycloakClientSecret string
	KeycloakJWKSURL      string

	SvedprintServiceURL      string
	SvedprintAdminServiceURL string
	SvedprintPrintServiceURL string
	GatewayDatabaseURL       string

	LogLevel string
}

func Load(serviceName string) (*Config, error) {
	cfg := &Config{
		ServiceName: serviceName,
		Port:        getEnv("PORT", "8080"),
		GinMode:     getEnv("GIN_MODE", "debug"),

		DatabaseURL:          getEnv("DATABASE_URL", ""),
		DatabaseMaxConns:     getEnvInt("DATABASE_MAX_CONNS", 25),
		DatabaseMaxIdleConns: getEnvInt("DATABASE_MAX_IDLE_CONNS", 10),
		DatabaseConnLifetime: getEnvDuration("DATABASE_CONN_MAX_LIFETIME", 5*time.Minute),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),
		RedisTTL:      getEnvDuration("REDIS_TTL", 10*time.Minute),

		KeycloakURL:          getEnv("KEYCLOAK_URL", "http://localhost:8080"),
		KeycloakRealm:        getEnv("KEYCLOAK_REALM", "svedprint"),
		KeycloakClientID:     getEnv("KEYCLOAK_CLIENT_ID", "svedprint-backend"),
		KeycloakClientSecret: getEnv("KEYCLOAK_CLIENT_SECRET", ""),
		KeycloakJWKSURL:      getEnv("KEYCLOAK_JWKS_URL", ""),

		SvedprintServiceURL:      getEnv("SVEDPRINT_SERVICE_URL", "http://svedprint:8001"),
		SvedprintAdminServiceURL: getEnv("SVEDPRINT_ADMIN_SERVICE_URL", "http://svedprint-admin:8002"),
		SvedprintPrintServiceURL: getEnv("SVEDPRINT_PRINT_SERVICE_URL", "http://svedprint-print:8003"),

		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	// Validate required fields based on service
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}

	// Service-specific validation
	switch c.ServiceName {
	case "gateway":
		if c.DatabaseURL == "" {
			return fmt.Errorf("DATABASE_URL is required for gateway service")
		}
		if c.KeycloakJWKSURL == "" {
			return fmt.Errorf("KEYCLOAK_JWKS_URL is required for gateway service")
		}
	case "svedprint", "svedprint-admin":
		if c.DatabaseURL == "" {
			return fmt.Errorf("DATABASE_URL is required for %s service", c.ServiceName)
		}
	case "svedprint-print":
		// Print service is stateless, no database required
	default:
		return fmt.Errorf("unknown service name: %s", c.ServiceName)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
