package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/chats/go-user-api/internal/domain/entity"
	"github.com/chats/go-user-api/internal/infrastructure/cache"
	"github.com/chats/go-user-api/internal/infrastructure/db"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

const userCacheKeyPrefix = "user:"
const userCacheTTL = 30 * time.Minute

// UserRepository defines the interface for user repository operations
type UserRepository interface {
	// Create a new user
	Create(ctx context.Context, user *entity.User) error

	// Get a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)

	// Get a user by email
	GetByEmail(ctx context.Context, email string) (*entity.User, error)

	// Get a user by username
	GetByUsername(ctx context.Context, username string) (*entity.User, error)

	// Update user information
	Update(ctx context.Context, user *entity.User) error

	// Delete a user
	Delete(ctx context.Context, id uuid.UUID) error

	// List users with pagination
	List(ctx context.Context, page, limit int) ([]*entity.User, int64, error)

	// Change user password
	ChangePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error

	// Update user status
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

type userRepository struct {
	db    db.Database
	cache cache.Cache
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db db.Database, cache cache.Cache) UserRepository {
	return &userRepository{
		db:    db,
		cache: cache,
	}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	// Get the appropriate instance based on the database type
	switch db := r.db.GetInstance().(type) {
	//case *pgxpool.Pool:
	//	return r.createUserPostgres(ctx, db, user)
	case *mongo.Client:
		return r.createUserMongo(ctx, db, user)
	default:
		return errors.New("unsupported database type")
	}
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("%s%s", userCacheKeyPrefix, id.String())
	cachedData, err := r.cache.Get(ctx, cacheKey)
	if err == nil && cachedData != nil {
		var user entity.User
		if err := json.Unmarshal(cachedData, &user); err == nil {
			return &user, nil
		}
		// If unmarshal fails, continue to get from database
	}

	// Get from database
	var user *entity.User
	var dbErr error

	switch db := r.db.GetInstance().(type) {
	//case *pgxpool.Pool:
	//	user, dbErr = r.getUserByIDPostgres(ctx, db, id)
	case *mongo.Client:
		user, dbErr = r.getUserByIDMongo(ctx, db, id)
	default:
		return nil, errors.New("unsupported database type")
	}

	if dbErr != nil {
		return nil, dbErr
	}

	// If user found, cache it
	if user != nil {
		if userData, err := json.Marshal(user); err == nil {
			if err := r.cache.Set(ctx, cacheKey, userData, userCacheTTL); err != nil {
				log.Warn().Err(err).Str("user_id", id.String()).Msg("Failed to cache user")
			}
		}
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	// Get from database
	switch db := r.db.GetInstance().(type) {
	//case *pgxpool.Pool:
	//	return r.getUserByEmailPostgres(ctx, db, email)
	case *mongo.Client:
		return r.getUserByEmailMongo(ctx, db, email)
	default:
		return nil, errors.New("unsupported database type")
	}
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	// Get from database
	switch db := r.db.GetInstance().(type) {
	//case *pgxpool.Pool:
	//	return r.getUserByUsernamePostgres(ctx, db, username)
	case *mongo.Client:
		return r.getUserByUsernameMongo(ctx, db, username)
	default:
		return nil, errors.New("unsupported database type")
	}
}

// Update updates user information
func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	// Update database
	var err error
	switch db := r.db.GetInstance().(type) {
	//case *pgxpool.Pool:
	//	err = r.updateUserPostgres(ctx, db, user)
	case *mongo.Client:
		err = r.updateUserMongo(ctx, db, user)
	default:
		return errors.New("unsupported database type")
	}

	if err != nil {
		return err
	}

	// Update cache
	cacheKey := fmt.Sprintf("%s%s", userCacheKeyPrefix, user.ID.String())
	if userData, err := json.Marshal(user); err == nil {
		if err := r.cache.Set(ctx, cacheKey, userData, userCacheTTL); err != nil {
			log.Warn().Err(err).Str("user_id", user.ID.String()).Msg("Failed to update user in cache")
		}
	}

	return nil
}

// Delete deletes a user
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {

	// Delete from database
	var err error
	switch db := r.db.GetInstance().(type) {
	//case *pgxpool.Pool:
	//	err = r.deleteUserPostgres(ctx, db, id)
	case *mongo.Client:
		err = r.deleteUserMongo(ctx, db, id)
	default:
		return errors.New("unsupported database type")
	}

	if err != nil {
		return err
	}

	// Delete from cache
	cacheKey := fmt.Sprintf("%s%s", userCacheKeyPrefix, id.String())
	if err := r.cache.Delete(ctx, cacheKey); err != nil {
		log.Warn().Err(err).Str("user_id", id.String()).Msg("Failed to delete user from cache")
	}

	return nil
}

// List retrieves a list of users with pagination
func (r *userRepository) List(ctx context.Context, page, limit int) ([]*entity.User, int64, error) {
	// Calculate offset
	offset := (page - 1) * limit

	// Get from database
	switch db := r.db.GetInstance().(type) {
	//case *pgxpool.Pool:
	//	return r.listUsersPostgres(ctx, db, limit, offset)
	case *mongo.Client:
		return r.listUsersMongo(ctx, db, limit, offset)
	default:
		return nil, 0, errors.New("unsupported database type")
	}
}

// ChangePassword changes a user's password
func (r *userRepository) ChangePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error {
	// Update database
	var err error
	switch db := r.db.GetInstance().(type) {
	//case *pgxpool.Pool:
	//	err = r.changePasswordPostgres(ctx, db, id, hashedPassword)
	case *mongo.Client:
		err = r.changePasswordMongo(ctx, db, id, hashedPassword)
	default:
		return errors.New("unsupported database type")
	}

	if err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("%s%s", userCacheKeyPrefix, id.String())
	if err := r.cache.Delete(ctx, cacheKey); err != nil {
		log.Warn().Err(err).Str("user_id", id.String()).Msg("Failed to invalidate user cache after password change")
	}

	return nil
}

// UpdateStatus updates a user's status
func (r *userRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	// Update database
	var err error
	switch db := r.db.GetInstance().(type) {
	//case *pgxpool.Pool:
	//	err = r.updateStatusPostgres(ctx, db, id, status)
	case *mongo.Client:
		err = r.updateStatusMongo(ctx, db, id, status)
	default:
		return errors.New("unsupported database type")
	}

	if err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("%s%s", userCacheKeyPrefix, id.String())
	if err := r.cache.Delete(ctx, cacheKey); err != nil {
		log.Warn().Err(err).Str("user_id", id.String()).Msg("Failed to invalidate user cache after status update")
	}

	return nil
}
