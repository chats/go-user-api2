package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/chats/go-user-api/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// createUserMongo creates a user in MongoDB
func (r *userRepository) createUserMongo(ctx context.Context, client *mongo.Client, user *entity.User) error {
	collection := client.Database("user_service").Collection("users")
	_, err := collection.InsertOne(ctx, user)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("Failed to create user in MongoDB")
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// getUserByIDMongo gets a user by ID from MongoDB
func (r *userRepository) getUserByIDMongo(ctx context.Context, client *mongo.Client, id uuid.UUID) (*entity.User, error) {
	collection := client.Database("user_service").Collection("users")

	var user entity.User
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // User not found
		}
		log.Error().Err(err).Str("user_id", id.String()).Msg("Failed to get user from MongoDB")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// getUserByEmailMongo gets a user by email from MongoDB
func (r *userRepository) getUserByEmailMongo(ctx context.Context, client *mongo.Client, email string) (*entity.User, error) {
	collection := client.Database("user_service").Collection("users")

	var user entity.User
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // User not found
		}
		log.Error().Err(err).Str("email", email).Msg("Failed to get user by email from MongoDB")
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// getUserByUsernameMongo gets a user by username from MongoDB
func (r *userRepository) getUserByUsernameMongo(ctx context.Context, client *mongo.Client, username string) (*entity.User, error) {
	collection := client.Database("user_service").Collection("users")

	var user entity.User
	err := collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // User not found
		}
		log.Error().Err(err).Str("username", username).Msg("Failed to get user by username from MongoDB")
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// updateUserMongo updates a user in MongoDB
func (r *userRepository) updateUserMongo(ctx context.Context, client *mongo.Client, user *entity.User) error {
	collection := client.Database("user_service").Collection("users")

	update := bson.M{
		"$set": bson.M{
			"email":      user.Email,
			"username":   user.Username,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
			"status":     user.Status,
			"updated_at": user.UpdatedAt,
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("Failed to update user in MongoDB")
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// deleteUserMongo deletes a user from MongoDB
func (r *userRepository) deleteUserMongo(ctx context.Context, client *mongo.Client, id uuid.UUID) error {
	collection := client.Database("user_service").Collection("users")

	_, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		log.Error().Err(err).Str("user_id", id.String()).Msg("Failed to delete user from MongoDB")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// listUsersMongo lists users from MongoDB
func (r *userRepository) listUsersMongo(ctx context.Context, client *mongo.Client, limit, offset int) ([]*entity.User, int64, error) {
	collection := client.Database("user_service").Collection("users")

	// Get total count
	total, countErr := collection.CountDocuments(ctx, bson.M{})
	if countErr != nil {
		log.Error().Err(countErr).Msg("Failed to count users in MongoDB")
		return nil, 0, fmt.Errorf("failed to count users: %w", countErr)
	}

	// Set options for pagination and sorting
	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	// Find users
	cursor, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list users from MongoDB")
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*entity.User
	if err := cursor.All(ctx, &users); err != nil {
		log.Error().Err(err).Msg("Failed to decode users from MongoDB")
		return nil, 0, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, total, nil
}

// changePasswordMongo changes a user's password in MongoDB
func (r *userRepository) changePasswordMongo(ctx context.Context, client *mongo.Client, id uuid.UUID, hashedPassword string) error {
	collection := client.Database("user_service").Collection("users")

	update := bson.M{
		"$set": bson.M{
			"password":   hashedPassword,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		log.Error().Err(err).Str("user_id", id.String()).Msg("Failed to change password in MongoDB")
		return fmt.Errorf("failed to change password: %w", err)
	}

	return nil
}

// updateStatusMongo updates a user's status in MongoDB
func (r *userRepository) updateStatusMongo(ctx context.Context, client *mongo.Client, id uuid.UUID, status string) error {
	collection := client.Database("user_service").Collection("users")

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		log.Error().Err(err).Str("user_id", id.String()).Msg("Failed to update status in MongoDB")
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}
