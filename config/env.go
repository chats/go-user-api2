package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// getEnv returns the value of the environment variable with fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvAsBool returns the boolean value of the environment variable with fallback
func getEnvAsBool(key string, fallback bool) bool {
	valStr := getEnv(key, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return fallback
}

// getEnvAsInt returns the integer value of the environment variable with fallback
func getEnvAsInt(key string, fallback int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return fallback
}

// getEnvAsFloat returns the float value of the environment variable with fallback
func getEnvAsFloat(key string, fallback float64) float64 {
	valStr := getEnv(key, "")
	if val, err := strconv.ParseFloat(valStr, 64); err == nil {
		return val
	}
	return fallback
}

// getEnvAsDuration returns the duration value of the environment variable with fallback
func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	valStr := getEnv(key, "")
	if val, err := time.ParseDuration(valStr); err == nil {
		return val
	}
	return fallback
}

// getEnvAsSlice returns the slice value of the environment variable with fallback
func getEnvAsSlice(key, sep string, fallback []string) []string {
	valStr := getEnv(key, "")
	if valStr == "" {
		return fallback
	}
	return strings.Split(valStr, sep)
}

// LoadEnv loads environment variables from .env file
func LoadEnv() {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Info().Msg(".env file not found, using environment variables")
	} else {
		log.Info().Msg("Loaded .env file")
	}
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	LoadEnv()

	// Create new config
	return &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "go-user-api"),
			Environment: getEnv("APP_ENV", "development"),
		},
		HTTP: HTTPConfig{
			Port:              getEnvAsInt("HTTP_PORT", 8080),
			ReadTimeout:       getEnvAsDuration("HTTP_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:      getEnvAsDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:       getEnvAsDuration("HTTP_IDLE_TIMEOUT", 120*time.Second),
			EnablePrefork:     getEnvAsBool("HTTP_ENABLE_PREFORK", false),
			EnableCompression: getEnvAsBool("HTTP_ENABLE_COMPRESSION", true),
		},
		GRPC: GRPCConfig{
			Port:             getEnvAsInt("GRPC_PORT", 50051),
			MaxRecvMsgSize:   getEnvAsInt("GRPC_MAX_RECV_MSG_SIZE", 4*1024*1024), // 4MB
			MaxSendMsgSize:   getEnvAsInt("GRPC_MAX_SEND_MSG_SIZE", 4*1024*1024), // 4MB
			EnableReflection: getEnvAsBool("GRPC_ENABLE_REFLECTION", true),
			UseTLS:           getEnvAsBool("GRPC_USE_TLS", false),
			CertFile:         getEnv("GRPC_CERT_FILE", ""),
			KeyFile:          getEnv("GRPC_KEY_FILE", ""),
		},
		Database: DatabaseConfig{
			Type:     DatabaseType(getEnv("DB_TYPE", "postgresql")),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			Username: getEnv("DB_USERNAME", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Database: getEnv("DB_DATABASE", "user_service"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Cache: CacheConfig{
			Type:     CacheType(getEnv("CACHE_TYPE", "redis")),
			Host:     getEnv("CACHE_HOST", "localhost"),
			Port:     getEnvAsInt("CACHE_PORT", 6379),
			Password: getEnv("CACHE_PASSWORD", ""),
			DB:       getEnvAsInt("CACHE_DB", 0),
		},
		Jaeger: JaegerConfig{
			Host:        getEnv("JAEGER_HOST", "localhost"),
			Port:        getEnvAsInt("JAEGER_PORT", 6831),
			ServiceName: getEnv("JAEGER_SERVICE_NAME", "go-user-api"),
			Enabled:     getEnvAsBool("JAEGER_ENABLED", true),
		},
		Security: SecurityConfig{
			JWTSecret:                    getEnv("JWT_SECRET", "your-secret-key"),
			JWTExpirationHours:           getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
			BcryptCost:                   getEnvAsInt("BCRYPT_COST", 12),
			PasetoPrivateKey:             getEnv("PASETO_PRIVATE_KEY", ""),
			PasetoPublicKey:              getEnv("PASETO_PUBLIC_KEY", ""),
			AccessTokenExpirationMinutes: getEnvAsInt("ACCESS_TOKEN_EXPIRATION_MINUTES", 15),
			RefreshTokenExpirationDays:   getEnvAsInt("REFRESH_TOKEN_EXPIRATION_DAYS", 7),
		},
		Middleware: MiddlewareConfig{
			EnableTracing:     getEnvAsBool("MIDDLEWARE_TRACING", false),
			EnableRequestID:   getEnvAsBool("MIDDLEWARE_REQUEST_ID", false),
			EnableRecover:     getEnvAsBool("MIDDLEWARE_RECOVER", false),
			EnableCORS:        getEnvAsBool("MIDDLEWARE_CORS", false),
			EnableHelmet:      getEnvAsBool("MIDDLEWARE_HELMET", false),
			EnableRateLimiter: getEnvAsBool("MIDDLEWARE_RATE_LIMITER", false),
			EnableETag:        getEnvAsBool("MIDDLEWARE_ETAG", false),
			EnableCompression: getEnvAsBool("MIDDLEWARE_COMPRESSION", false),
		},
	}
}
