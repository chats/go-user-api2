package middleware

import (
	"errors"
	"strings"

	"github.com/chats/go-user-api/internal/domain/service"
	"github.com/chats/go-user-api/internal/domain/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// AuthMiddleware creates a middleware to validate access tokens
func AuthMiddleware(authUseCase usecase.AuthUseCase) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header is required",
			})
		}

		// Check if the header has the correct format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization format, expected 'Bearer {token}'",
			})
		}

		// Extract token
		token := parts[1]

		// Validate token
		userID, err := authUseCase.ValidateToken(c.Context(), token)
		if err != nil {
			log.Error().Err(err).Msg("Failed to validate token")

			if errors.Is(err, service.ErrInvalidToken) || errors.Is(err, service.ErrExpiredToken) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Invalid or expired token",
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to validate token",
			})
		}

		// Set user ID in context for later use
		c.Locals("user_id", userID)

		// In a real implementation, you would extract the token ID from the claims as well
		// For now we'll set a placeholder
		c.Locals("token_id", uuid.New())

		return c.Next()
	}
}

// RoleMiddleware creates a middleware to check user roles
func RoleMiddleware(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// This would typically extract the user role from the token or database
		// For simplicity, we're just checking if the role was set in the context

		// In a real implementation, you would get the user from the database or token claims
		// and check their role
		role, ok := c.Locals("user_role").(string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}

		// Check if the user has one of the required roles
		for _, r := range roles {
			if r == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Insufficient permissions",
		})
	}
}
