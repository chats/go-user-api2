package db

import (
	"context"
	"fmt"
	"time"

	"github.com/chats/go-user-api/config"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDatabase implements the Database interface for MongoDB
type MongoDatabase struct {
	config   config.DatabaseConfig
	client   *mongo.Client
	database *mongo.Database
}

// NewMongoDB creates a new MongoDB database connection
func NewMongoDB(config config.DatabaseConfig) (Database, error) {
	return &MongoDatabase{
		config: config,
	}, nil
}

// Connect establishes a connection to MongoDB
func (db *MongoDatabase) Connect(ctx context.Context) error {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d",
		db.config.Username,
		db.config.Password,
		db.config.Host,
		db.config.Port,
	)

	// Configure client options
	clientOptions := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(10 * time.Second).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(30 * time.Minute)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Ping the MongoDB server to verify connection
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("failed to ping MongoDB server: %v", err)
	}

	db.client = client
	db.database = client.Database(db.config.Database)
	log.Info().Msg("Connected to MongoDB successfully")
	return nil
}

// Close closes the MongoDB connection
func (db *MongoDatabase) Close(ctx context.Context) error {
	if db.client != nil {
		log.Info().Msg("Closing MongoDB connection")
		return db.client.Disconnect(ctx)
	}
	return nil
}

// Ping verifies the connection to MongoDB
func (db *MongoDatabase) Ping(ctx context.Context) error {
	if db.client == nil {
		return fmt.Errorf("MongoDB client not initialized")
	}
	return db.client.Ping(ctx, readpref.Primary())
}

// GetInstance returns the MongoDB client instance
func (db *MongoDatabase) GetInstance() interface{} {
	return db.client
}

// GetClient returns the MongoDB client
func (db *MongoDatabase) GetClient() *mongo.Client {
	return db.client
}

// GetDatabase returns the MongoDB database
func (db *MongoDatabase) GetDatabase() *mongo.Database {
	return db.database
}

// Collection returns a specific collection
func (db *MongoDatabase) Collection(name string) *mongo.Collection {
	return db.database.Collection(name)
}
