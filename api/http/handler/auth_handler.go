package handler

import (
	"errors"

	"github.com/chats/go-user-api/internal/domain/entity"
	"github.com/chats/go-user-api/internal/domain/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	authUseCase usecase.AuthUseCase
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authUseCase usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

// RegisterRoutes registers the routes for the auth handler
func (h *AuthHandler) RegisterRoutes(router fiber.Router, authMiddleware fiber.Handler) {
	authGroup := router.Group("/auth")

	// Public routes
	authGroup.Post("/login", h.Login)
	authGroup.Post("/refresh", h.RefreshToken)

	// Protected routes
	authGroup.Post("/logout", authMiddleware, h.Logout)
	authGroup.Post("/logout-all", authMiddleware, h.LogoutAll)
}

// Login handles user login and returns access and refresh tokens
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	// Parse request body
	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse login request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}

	// Login user
	response, err := h.authUseCase.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("Failed to login user")

		if errors.Is(err, usecase.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to login user",
		})
	}

	// Return tokens and user info
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user": fiber.Map{
			"id":         response.User.ID,
			"email":      response.User.Email,
			"username":   response.User.Username,
			"first_name": response.User.FirstName,
			"last_name":  response.User.LastName,
			"role":       response.User.Role,
			"status":     response.User.Status,
		},
		"token_type":    "Bearer",
		"access_token":  response.AuthTokens.AccessToken,
		"refresh_token": response.AuthTokens.RefreshToken,
		"expires_at":    response.AuthTokens.ExpiresAt,
	})
}

// RefreshToken refreshes the access token using a refresh token
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	// Parse request body
	var req entity.RefreshTokenRequest

	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse refresh token request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Refresh token is required",
		})
	}

	// Refresh token
	tokens, err := h.authUseCase.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		log.Error().Err(err).Msg("Failed to refresh token")

		if errors.Is(err, usecase.ErrInvalidRefreshToken) || errors.Is(err, usecase.ErrRefreshTokenExpired) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired refresh token",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to refresh token",
		})
	}

	// Return new tokens
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"token_type":    "Bearer",
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_at":    tokens.ExpiresAt,
	})
}

// Logout logs out a user by invalidating their access token
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Get token ID from context
	tokenID, ok := c.Locals("token_id").(uuid.UUID)
	if !ok {
		log.Error().Msg("Token ID not found in context")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to logout",
		})
	}

	// Logout user
	if err := h.authUseCase.Logout(c.Context(), tokenID); err != nil {
		log.Error().Err(err).Str("token_id", tokenID.String()).Msg("Failed to logout user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to logout",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Successfully logged out",
	})
}

// LogoutAll logs out a user from all devices
func (h *AuthHandler) LogoutAll(c *fiber.Ctx) error {
	// Get user ID from context
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		log.Error().Msg("User ID not found in context")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to logout from all devices",
		})
	}

	// Logout user from all devices
	if err := h.authUseCase.LogoutAll(c.Context(), userID); err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to logout user from all devices")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to logout from all devices",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Successfully logged out from all devices",
	})
}
