package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/chats/go-user-api/internal/domain/entity"
	"github.com/chats/go-user-api/internal/domain/repository"
	"github.com/chats/go-user-api/internal/domain/service"
	"github.com/chats/go-user-api/utils"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var (
	// ErrInvalidRefreshToken is returned when a refresh token is invalid
	ErrInvalidRefreshToken = errors.New("invalid refresh token")

	// ErrRefreshTokenExpired is returned when a refresh token is expired
	ErrRefreshTokenExpired = errors.New("refresh token expired")
)

// AuthUseCase defines the use case for authentication operations
type AuthUseCase interface {
	// Login authenticates a user and returns tokens
	Login(ctx context.Context, email, password string) (*entity.LoginResponse, error)

	// Logout invalidates a user's tokens
	Logout(ctx context.Context, tokenID uuid.UUID) error

	// RefreshToken refreshes the access token using a refresh token
	RefreshToken(ctx context.Context, refreshToken string) (*entity.AuthTokens, error)

	// LogoutAll invalidates all of a user's tokens
	LogoutAll(ctx context.Context, userID uuid.UUID) error

	// ValidateToken validates a token and returns the user ID
	ValidateToken(ctx context.Context, token string) (uuid.UUID, error)
}

type authUseCase struct {
	userRepo     repository.UserRepository
	tokenRepo    repository.TokenRepository
	tokenService service.TokenService
}

// NewAuthUseCase creates a new AuthUseCase
func NewAuthUseCase(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	tokenService service.TokenService,
) AuthUseCase {
	return &authUseCase{
		userRepo:     userRepo,
		tokenRepo:    tokenRepo,
		tokenService: tokenService,
	}
}

// Login authenticates a user and returns tokens
func (uc *authUseCase) Login(ctx context.Context, email, password string) (*entity.LoginResponse, error) {
	// Authenticate user
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password - using the utils function
	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	tokens, accessDetails, refreshDetails, err := uc.tokenService.GenerateTokens(user.ID)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("Failed to generate tokens")
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Store tokens in Redis
	if err := uc.tokenRepo.StoreAccessToken(ctx, accessDetails); err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("Failed to store access token")
		return nil, fmt.Errorf("failed to store access token: %w", err)
	}

	if err := uc.tokenRepo.StoreRefreshToken(ctx, refreshDetails); err != nil {
		log.Error().Err(err).Str("user_id", user.ID.String()).Msg("Failed to store refresh token")
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &entity.LoginResponse{
		User:       user,
		AuthTokens: *tokens,
	}, nil
}

// Logout invalidates a user's token
func (uc *authUseCase) Logout(ctx context.Context, tokenID uuid.UUID) error {
	// Delete access token
	if err := uc.tokenRepo.DeleteToken(ctx, tokenID, entity.AccessToken); err != nil {
		log.Error().Err(err).Str("token_id", tokenID.String()).Msg("Failed to delete access token")
		return fmt.Errorf("failed to delete access token: %w", err)
	}

	return nil
}

// RefreshToken refreshes the access token using a refresh token
func (uc *authUseCase) RefreshToken(ctx context.Context, refreshToken string) (*entity.AuthTokens, error) {
	// Validate refresh token
	claims, err := uc.tokenService.ValidateToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Verify it's a refresh token
	if claims.TokenType != entity.RefreshToken {
		return nil, ErrInvalidRefreshToken
	}

	// Get token from Redis to verify it hasn't been revoked
	tokenDetails, err := uc.tokenRepo.GetToken(ctx, claims.TokenID, entity.RefreshToken)
	if err != nil {
		log.Error().Err(err).Str("token_id", claims.TokenID.String()).Msg("Failed to get refresh token")
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	if tokenDetails == nil {
		return nil, ErrInvalidRefreshToken
	}

	// Generate new tokens
	tokens, accessDetails, refreshDetails, err := uc.tokenService.GenerateTokens(claims.UserID)
	if err != nil {
		log.Error().Err(err).Str("user_id", claims.UserID.String()).Msg("Failed to generate new tokens")
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	// Store new tokens in Redis
	if err := uc.tokenRepo.StoreAccessToken(ctx, accessDetails); err != nil {
		log.Error().Err(err).Str("user_id", claims.UserID.String()).Msg("Failed to store new access token")
		return nil, fmt.Errorf("failed to store new access token: %w", err)
	}

	if err := uc.tokenRepo.StoreRefreshToken(ctx, refreshDetails); err != nil {
		log.Error().Err(err).Str("user_id", claims.UserID.String()).Msg("Failed to store new refresh token")
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	// Delete old refresh token
	if err := uc.tokenRepo.DeleteToken(ctx, claims.TokenID, entity.RefreshToken); err != nil {
		log.Warn().Err(err).Str("token_id", claims.TokenID.String()).Msg("Failed to delete old refresh token")
	}

	return tokens, nil
}

// LogoutAll invalidates all of a user's tokens
func (uc *authUseCase) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	// Delete all user tokens from Redis
	if err := uc.tokenRepo.DeleteUserTokens(ctx, userID); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to delete all user tokens")
		return fmt.Errorf("failed to delete all user tokens: %w", err)
	}

	return nil
}

// ValidateToken validates a token and returns the user ID
func (uc *authUseCase) ValidateToken(ctx context.Context, token string) (uuid.UUID, error) {
	// Validate token
	claims, err := uc.tokenService.ValidateToken(token)
	if err != nil {
		return uuid.Nil, service.ErrInvalidToken
	}

	// Verify it's an access token
	if claims.TokenType != entity.AccessToken {
		return uuid.Nil, service.ErrInvalidToken
	}

	// Get token from Redis to verify it hasn't been revoked
	tokenDetails, err := uc.tokenRepo.GetToken(ctx, claims.TokenID, entity.AccessToken)
	if err != nil {
		log.Error().Err(err).Str("token_id", claims.TokenID.String()).Msg("Failed to get access token")
		return uuid.Nil, fmt.Errorf("failed to get access token: %w", err)
	}

	if tokenDetails == nil {
		return uuid.Nil, service.ErrInvalidToken
	}

	return claims.UserID, nil
}
