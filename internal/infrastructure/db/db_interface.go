package db

import "context"

// Database defines the interface for database operations
type Database interface {
	// Connect establishes a connection to the database
	Connect(ctx context.Context) error

	// Close closes the database connection
	Close(ctx context.Context) error

	// Ping verifies the connection to the database
	Ping(ctx context.Context) error

	// GetInstance returns the database instance
	GetInstance() interface{}
}
