package config

import (
	"time"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	// PostgreSQL database type
	PostgreSQL DatabaseType = "postgresql"
	// MongoDB database type
	MongoDB DatabaseType = "mongodb"
)

// CacheType represents the type of cache
type CacheType string

const (
	// Redis cache type
	Redis CacheType = "redis"
	// Memcached cache type
	Memcached CacheType = "memcached"
)

// Config contains all application configuration
type Config struct {
	App        AppConfig
	HTTP       HTTPConfig
	GRPC       GRPCConfig
	Database   DatabaseConfig
	Cache      CacheConfig
	Jaeger     JaegerConfig
	Security   SecurityConfig
	Middleware MiddlewareConfig
}

// AppConfig contains general application configuration
type AppConfig struct {
	Name        string
	Environment string
}

// HTTPConfig contains HTTP server configuration
type HTTPConfig struct {
	Port              int
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	EnablePrefork     bool
	EnableCompression bool
}

// GRPCConfig contains gRPC server configuration
type GRPCConfig struct {
	Port             int
	MaxRecvMsgSize   int
	MaxSendMsgSize   int
	EnableReflection bool
	UseTLS           bool
	CertFile         string
	KeyFile          string
}

// CacheConfig contains cache configuration
type CacheConfig struct {
	Type     CacheType
	Host     string
	Port     int
	Password string
	DB       int // For Redis
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	Type     DatabaseType
	Host     string
	Port     int
	Username string
	Password string
	Database string
	SSLMode  string
}

// JaegerConfig contains Jaeger configuration
type JaegerConfig struct {
	Host        string
	Port        int
	ServiceName string
	Enabled     bool
}

// SecurityConfig contains security configuration
type SecurityConfig struct {
	JWTSecret          string
	JWTExpirationHours int
	BcryptCost         int

	// PASETO related fields
	PasetoPrivateKey string
	PasetoPublicKey  string

	// Token expiration settings
	AccessTokenExpirationMinutes int
	RefreshTokenExpirationDays   int
}

type MiddlewareConfig struct {
	EnableTracing     bool
	EnableRequestID   bool
	EnableRecover     bool
	EnableCORS        bool
	EnableHelmet      bool
	EnableRateLimiter bool
	EnableETag        bool
	EnableCompression bool
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// IsStaging returns true if the environment is staging
func (c *Config) IsStaging() bool {
	return c.App.Environment == "staging"
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development" || c.App.Environment == "local"
}
