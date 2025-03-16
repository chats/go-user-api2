package db

import (
	"fmt"

	"github.com/chats/go-user-api/config"
	"github.com/rs/zerolog/log"
)

// Factory is an interface for creating database connections
type Factory interface {
	Create(config config.DatabaseConfig) (Database, error)
}

// DatabaseFactory implements the Factory interface
type DatabaseFactory struct{}

// NewDatabaseFactory creates a new DatabaseFactory
func NewDatabaseFactory() Factory {
	return &DatabaseFactory{}
}

// Create creates a new database connection based on the provided configuration
func (f *DatabaseFactory) Create(config config.DatabaseConfig) (Database, error) {
	switch config.Type {
	//case "postgresql":
	//	log.Info().Msg("Creating PostgreSQL database connection")
	//	return NewPostgreSQL(config)
	case "mongodb":
		log.Info().Msg("Creating MongoDB database connection")
		return NewMongoDB(config)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}
