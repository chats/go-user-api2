package entity

import (
	"time"

	"github.com/google/uuid"
)

// TokenType defines the type of token
type TokenType string

const (
	// AccessToken represents an access token
	AccessToken TokenType = "access"
	// RefreshToken represents a refresh token
	RefreshToken TokenType = "refresh"
)

// TokenDetails contains the metadata of a token
type TokenDetails struct {
	TokenID    uuid.UUID `json:"token_id"`
	UserID     uuid.UUID `json:"user_id"`
	TokenType  TokenType `json:"token_type"`
	Expiration time.Time `json:"expiration"`
}

// AuthTokens contains both access and refresh tokens
type AuthTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// RefreshTokenRequest is used for refresh token requests
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LoginResponse is the response for login requests
type LoginResponse struct {
	User       *User      `json:"user"`
	AuthTokens AuthTokens `json:"auth_tokens"`
}
