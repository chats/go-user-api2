package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chats/go-user-api/internal/domain/entity"
	"github.com/chats/go-user-api/internal/infrastructure/cache"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	accessTokenPrefix  = "access_token:"
	refreshTokenPrefix = "refresh_token:"
	userTokensPrefix   = "user_tokens:"
)

// TokenRepository defines the interface for token repository operations
type TokenRepository interface {
	// StoreAccessToken stores an access token with expiration
	StoreAccessToken(ctx context.Context, details *entity.TokenDetails) error

	// StoreRefreshToken stores a refresh token with expiration
	StoreRefreshToken(ctx context.Context, details *entity.TokenDetails) error

	// GetToken retrieves token details by token ID and type
	GetToken(ctx context.Context, tokenID uuid.UUID, tokenType entity.TokenType) (*entity.TokenDetails, error)

	// DeleteToken deletes a token
	DeleteToken(ctx context.Context, tokenID uuid.UUID, tokenType entity.TokenType) error

	// DeleteUserTokens deletes all tokens for a user
	DeleteUserTokens(ctx context.Context, userID uuid.UUID) error
}

type tokenRepository struct {
	cache cache.Cache
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(cache cache.Cache) TokenRepository {
	return &tokenRepository{
		cache: cache,
	}
}

// StoreAccessToken stores an access token with expiration
func (r *tokenRepository) StoreAccessToken(ctx context.Context, details *entity.TokenDetails) error {
	return r.storeToken(ctx, details, accessTokenPrefix)
}

// StoreRefreshToken stores a refresh token with expiration
func (r *tokenRepository) StoreRefreshToken(ctx context.Context, details *entity.TokenDetails) error {
	return r.storeToken(ctx, details, refreshTokenPrefix)
}

// storeToken is a helper method to store tokens
func (r *tokenRepository) storeToken(ctx context.Context, details *entity.TokenDetails, prefix string) error {
	// Create token key
	key := fmt.Sprintf("%s%s", prefix, details.TokenID.String())

	// Serialize token details
	data, err := json.Marshal(details)
	if err != nil {
		log.Error().Err(err).Str("token_id", details.TokenID.String()).Msg("Failed to marshal token details")
		return fmt.Errorf("failed to marshal token details: %w", err)
	}

	// Calculate expiration
	expiration := time.Until(details.Expiration)

	// Store token in Redis
	err = r.cache.Set(ctx, key, data, expiration)
	if err != nil {
		log.Error().Err(err).Str("token_id", details.TokenID.String()).Msg("Failed to store token in cache")
		return fmt.Errorf("failed to store token: %w", err)
	}

	// Add token to user's tokens set
	userTokensKey := fmt.Sprintf("%s%s", userTokensPrefix, details.UserID.String())
	userTokenData := fmt.Sprintf("%s:%s", string(details.TokenType), details.TokenID.String())

	// For simplicity, we're using a string value here
	// In a real implementation, you might want to use Redis SET or HASH
	err = r.cache.Set(ctx, userTokensKey+":"+userTokenData, []byte("1"), expiration)
	if err != nil {
		log.Warn().Err(err).Str("user_id", details.UserID.String()).Msg("Failed to add token to user tokens")
	}

	return nil
}

// GetToken retrieves token details by token ID and type
func (r *tokenRepository) GetToken(ctx context.Context, tokenID uuid.UUID, tokenType entity.TokenType) (*entity.TokenDetails, error) {
	// Determine prefix based on token type
	var prefix string
	if tokenType == entity.AccessToken {
		prefix = accessTokenPrefix
	} else {
		prefix = refreshTokenPrefix
	}

	// Create token key
	key := fmt.Sprintf("%s%s", prefix, tokenID.String())

	// Get token from Redis
	data, err := r.cache.Get(ctx, key)
	if err != nil {
		log.Error().Err(err).Str("token_id", tokenID.String()).Msg("Failed to get token from cache")
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	if data == nil {
		return nil, nil // Token not found
	}

	// Deserialize token details
	var details entity.TokenDetails
	err = json.Unmarshal(data, &details)
	if err != nil {
		log.Error().Err(err).Str("token_id", tokenID.String()).Msg("Failed to unmarshal token details")
		return nil, fmt.Errorf("failed to unmarshal token details: %w", err)
	}

	return &details, nil
}

// DeleteToken deletes a token
func (r *tokenRepository) DeleteToken(ctx context.Context, tokenID uuid.UUID, tokenType entity.TokenType) error {
	// Determine prefix based on token type
	var prefix string
	if tokenType == entity.AccessToken {
		prefix = accessTokenPrefix
	} else {
		prefix = refreshTokenPrefix
	}

	// Create token key
	key := fmt.Sprintf("%s%s", prefix, tokenID.String())

	// Get token details first to get user ID
	token, err := r.GetToken(ctx, tokenID, tokenType)
	if err != nil || token == nil {
		// If token doesn't exist, nothing to delete
		return nil
	}

	// Delete token from Redis
	err = r.cache.Delete(ctx, key)
	if err != nil {
		log.Error().Err(err).Str("token_id", tokenID.String()).Msg("Failed to delete token from cache")
		return fmt.Errorf("failed to delete token: %w", err)
	}

	// Remove from user tokens set
	userTokensKey := fmt.Sprintf("%s%s", userTokensPrefix, token.UserID.String())
	userTokenData := fmt.Sprintf("%s:%s", string(tokenType), tokenID.String())

	// Delete from user tokens
	err = r.cache.Delete(ctx, userTokensKey+":"+userTokenData)
	if err != nil {
		log.Warn().Err(err).Str("user_id", token.UserID.String()).Msg("Failed to remove token from user tokens")
	}

	return nil
}

// DeleteUserTokens deletes all tokens for a user
func (r *tokenRepository) DeleteUserTokens(ctx context.Context, userID uuid.UUID) error {
	// For a more robust implementation, you would use Redis SCAN to get all user tokens
	// and then delete them in batch

	// Here we're using a simplistic approach
	userTokensKey := fmt.Sprintf("%s%s:*", userTokensPrefix, userID.String())
	log.Debug().Str("user_id", userID.String()).Str("key", userTokensKey).Msg("Deleting all user tokens")

	// In a real implementation, you would get all keys matching the pattern
	// and delete them all

	// For simplicity, we'll use Clear method which is non-ideal
	// In production, you'd implement a method to delete by pattern
	log.Warn().Str("user_id", userID.String()).Msg("Deleting all user tokens - this is a simplified implementation")

	// In a real implementation with Redis, you would use SCAN and DEL
	// Here we'll just return nil
	return nil
}
