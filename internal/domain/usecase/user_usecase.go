package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/chats/go-user-api/internal/domain/entity"
	"github.com/chats/go-user-api/internal/domain/repository"
	"github.com/chats/go-user-api/utils"
	"github.com/google/uuid"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
)

// UserUseCase defines the use case for user operations
type UserUseCase interface {
	// Register creates a new user
	Register(ctx context.Context, email, username, password, firstName, lastName string) (*entity.User, error)

	// Get a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)

	// Update user information
	Update(ctx context.Context, id uuid.UUID, firstName, lastName string) (*entity.User, error)

	// Delete a user
	Delete(ctx context.Context, id uuid.UUID) error

	// List users with pagination
	List(ctx context.Context, page, limit int) ([]*entity.User, int64, error)

	// Change user password
	ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error

	// Update user status
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error

	// Authenticate user and return user if successful
	Authenticate(ctx context.Context, email, password string) (*entity.User, error)
}

// userUseCase implements UserUseCase interface
type userUseCase struct {
	userRepo repository.UserRepository
}

// NewUserUseCase creates a new UserUseCase
func NewUserUseCase(userRepo repository.UserRepository) UserUseCase {
	return &userUseCase{
		userRepo: userRepo,
	}
}

// Register creates a new user
func (uc *userUseCase) Register(ctx context.Context, email, username, password, firstName, lastName string) (*entity.User, error) {
	// Check if email already exists
	existingUser, err := uc.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Check if username already exists
	existingUser, err = uc.userRepo.GetByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return nil, ErrUsernameAlreadyExists
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := entity.NewUser(email, username, hashedPassword, firstName, lastName)

	// Save to repository
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (uc *userUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// Update updates a user's information
func (uc *userUseCase) Update(ctx context.Context, id uuid.UUID, firstName, lastName string) (*entity.User, error) {
	// Get user
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Update fields
	user.FirstName = firstName
	user.LastName = lastName
	user.UpdatedAt = time.Now()

	// Save changes
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Delete deletes a user
func (uc *userUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if user exists
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	return uc.userRepo.Delete(ctx, id)
}

// List lists users with pagination
func (uc *userUseCase) List(ctx context.Context, page, limit int) ([]*entity.User, int64, error) {
	return uc.userRepo.List(ctx, page, limit)
}

// ChangePassword changes a user's password
func (uc *userUseCase) ChangePassword(ctx context.Context, id uuid.UUID, oldPassword, newPassword string) error {
	// Get user
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verify old password
	if !utils.CheckPasswordHash(oldPassword, user.Password) {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return uc.userRepo.ChangePassword(ctx, id, hashedPassword)
}

// UpdateStatus updates a user's status
func (uc *userUseCase) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	// Check if user exists
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Validate status
	if status != entity.UserStatusActive &&
		status != entity.UserStatusInactive &&
		status != entity.UserStatusBlocked {
		return errors.New("invalid status")
	}

	return uc.userRepo.UpdateStatus(ctx, id, status)
}

// Authenticate authenticates a user
func (uc *userUseCase) Authenticate(ctx context.Context, email, password string) (*entity.User, error) {
	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if user.Status != entity.UserStatusActive {
		return nil, errors.New("user account is not active")
	}

	// Verify password
	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
