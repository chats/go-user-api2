package service

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/chats/go-user-api/config"
	"github.com/chats/go-user-api/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/o1egl/paseto"
)

var (
	// ErrInvalidToken is returned when a token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when a token is expired
	ErrExpiredToken = errors.New("token is expired")
)

// TokenClaims represents the claims in a token
type TokenClaims struct {
	TokenID   uuid.UUID        `json:"jti"`
	UserID    uuid.UUID        `json:"sub"`
	TokenType entity.TokenType `json:"type"`
}

// TokenService handles token operations
type TokenService interface {
	// GenerateTokens generates new access and refresh tokens
	GenerateTokens(userID uuid.UUID) (*entity.AuthTokens, *entity.TokenDetails, *entity.TokenDetails, error)

	// ValidateToken validates a token and returns its claims
	ValidateToken(token string) (*TokenClaims, error)

	// GetPublicKey returns the public key for token verification
	GetPublicKey() []byte
}

type tokenService struct {
	secretKey       string
	publicKey       ed25519.PublicKey
	privateKey      ed25519.PrivateKey
	accessDuration  time.Duration
	refreshDuration time.Duration
}

// NewTokenService creates a new token service
func NewTokenService(cfg config.SecurityConfig) (TokenService, error) {
	// Convert hex-encoded keys to byte slices
	privateKeyBytes, err := hex.DecodeString(cfg.PasetoPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	// For Ed25519, the private key contains the public key in the second half
	privateKey := ed25519.PrivateKey(privateKeyBytes)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	return &tokenService{
		secretKey:       cfg.JWTSecret,
		publicKey:       publicKey,
		privateKey:      privateKey,
		accessDuration:  time.Duration(cfg.AccessTokenExpirationMinutes) * time.Minute,
		refreshDuration: time.Duration(cfg.RefreshTokenExpirationDays) * 24 * time.Hour,
	}, nil
}

// GenerateTokens generates new access and refresh tokens
func (s *tokenService) GenerateTokens(userID uuid.UUID) (*entity.AuthTokens, *entity.TokenDetails, *entity.TokenDetails, error) {
	// Create token details
	accessTokenDetails := &entity.TokenDetails{
		TokenID:    uuid.New(),
		UserID:     userID,
		TokenType:  entity.AccessToken,
		Expiration: time.Now().Add(s.accessDuration),
	}

	refreshTokenDetails := &entity.TokenDetails{
		TokenID:    uuid.New(),
		UserID:     userID,
		TokenType:  entity.RefreshToken,
		Expiration: time.Now().Add(s.refreshDuration),
	}

	// Create new PASETO tokens
	accessToken, err := s.createToken(accessTokenDetails)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create access token: %w", err)
	}

	refreshToken, err := s.createToken(refreshTokenDetails)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &entity.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessTokenDetails.Expiration,
	}, accessTokenDetails, refreshTokenDetails, nil
}

// createToken creates a new PASETO token
func (s *tokenService) createToken(details *entity.TokenDetails) (string, error) {
	// Create a new PASETO token (v2.local for symmetric encryption or v2.public for asymmetric)
	v2 := paseto.NewV2()

	// Create footer (optional)
	footer := map[string]interface{}{
		"kid": "key-1", // Key ID for key rotation
	}

	// Create claims
	claims := TokenClaims{
		TokenID:   details.TokenID,
		UserID:    details.UserID,
		TokenType: details.TokenType,
	}

	// Sign token with claims
	// For v2.public we use asymmetric encryption (ed25519)
	token, err := v2.Sign(s.privateKey, claims, footer)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return token, nil
}

// ValidateToken validates a token and returns its claims
func (s *tokenService) ValidateToken(token string) (*TokenClaims, error) {
	v2 := paseto.NewV2()
	var claims TokenClaims
	var footer map[string]interface{}

	// Verify token and extract claims
	err := v2.Verify(token, s.publicKey, &claims, &footer)
	if err != nil {
		return nil, ErrInvalidToken
	}

	return &claims, nil
}

// GetPublicKey returns the public key for token verification
func (s *tokenService) GetPublicKey() []byte {
	return s.publicKey
}
