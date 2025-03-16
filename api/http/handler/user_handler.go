package handler

import (
	"errors"
	"time"

	"github.com/chats/go-user-api/internal/domain/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userUseCase usecase.UserUseCase
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userUseCase usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

// RegisterRoutes registers the routes for the user handler
func (h *UserHandler) RegisterRoutes(router fiber.Router) {
	userGroup := router.Group("/users")

	// Routes that don't require authentication
	userGroup.Post("/register", h.Register)
	userGroup.Post("/login", h.Login)

	// Routes that require authentication
	// In a real application, these would be protected by middleware
	userGroup.Get("/:id", h.GetByID)
	userGroup.Put("/:id", h.Update)
	userGroup.Delete("/:id", h.Delete)
	userGroup.Get("/", h.List)
	userGroup.Put("/:id/password", h.ChangePassword)
	userGroup.Put("/:id/status", h.UpdateStatus)
}

// Register handles user registration
func (h *UserHandler) Register(c *fiber.Ctx) error {
	// Parse request body
	var req struct {
		Email     string `json:"email" validate:"required,email"`
		Username  string `json:"username" validate:"required,min=3,max=50"`
		Password  string `json:"password" validate:"required,min=8"`
		FirstName string `json:"first_name" validate:"required"`
		LastName  string `json:"last_name" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse register request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	//span.SetAttributes(
	//		attribute.String("user.email", req.Email),
	//		attribute.String("user.username", req.Username),
	//	)

	// Validate request
	if req.Email == "" || req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email, username, and password are required",
		})
	}

	// Register user
	user, err := h.userUseCase.Register(c.Context(), req.Email, req.Username, req.Password, req.FirstName, req.LastName)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("Failed to register user")

		switch {
		case errors.Is(err, usecase.ErrEmailAlreadyExists):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Email already exists",
			})
		case errors.Is(err, usecase.ErrUsernameAlreadyExists):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Username already exists",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to register user",
			})
		}
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"role":       user.Role,
		"status":     user.Status,
		"created_at": user.CreatedAt,
	})
}

// Login handles user authentication
func (h *UserHandler) Login(c *fiber.Ctx) error {
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

	// Authenticate user
	user, err := h.userUseCase.Authenticate(c.Context(), req.Email, req.Password)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("Failed to authenticate user")

		if errors.Is(err, usecase.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to authenticate user",
		})
	}

	// In a real application, you would generate a JWT token here
	// For now, we'll just return the user information
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"role":       user.Role,
		"status":     user.Status,
		// Don't include the password in the response
	})
}

// GetByID gets a user by ID
func (h *UserHandler) GetByID(c *fiber.Ctx) error {
	// Parse user ID from path
	idParam := c.Params("id")
	if idParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	// Parse UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		log.Error().Err(err).Str("id", idParam).Msg("Invalid user ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	// Get user
	user, err := h.userUseCase.GetByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Str("id", idParam).Msg("Failed to get user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user",
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Return user
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"role":       user.Role,
		"status":     user.Status,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
}

// Update updates a user
func (h *UserHandler) Update(c *fiber.Ctx) error {
	// Parse user ID from path
	idParam := c.Params("id")
	if idParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	// Parse UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		log.Error().Err(err).Str("id", idParam).Msg("Invalid user ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	// Parse request body
	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse update request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update user
	user, err := h.userUseCase.Update(c.Context(), id, req.FirstName, req.LastName)
	if err != nil {
		log.Error().Err(err).Str("id", idParam).Msg("Failed to update user")

		if errors.Is(err, usecase.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user",
		})
	}

	// Return updated user
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"role":       user.Role,
		"status":     user.Status,
		"updated_at": user.UpdatedAt,
	})
}

// Delete deletes a user
func (h *UserHandler) Delete(c *fiber.Ctx) error {
	// Parse user ID from path
	idParam := c.Params("id")
	if idParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	// Parse UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		log.Error().Err(err).Str("id", idParam).Msg("Invalid user ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	// Delete user
	err = h.userUseCase.Delete(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Str("id", idParam).Msg("Failed to delete user")

		if errors.Is(err, usecase.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete user",
		})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}

// List lists users with pagination
func (h *UserHandler) List(c *fiber.Ctx) error {
	// Parse pagination params
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}

	limit := c.QueryInt("limit", 10)
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// List users
	users, total, err := h.userUseCase.List(c.Context(), page, limit)
	if err != nil {
		log.Error().Err(err).Int("page", page).Int("limit", limit).Msg("Failed to list users")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list users",
		})
	}

	// Map users to response format
	userResponses := make([]fiber.Map, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, fiber.Map{
			"id":         user.ID,
			"email":      user.Email,
			"username":   user.Username,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
			"status":     user.Status,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		})
	}

	// Return users
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"users": userResponses,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// ChangePassword changes a user's password
func (h *UserHandler) ChangePassword(c *fiber.Ctx) error {
	// Parse user ID from path
	idParam := c.Params("id")
	if idParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	// Parse UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		log.Error().Err(err).Str("id", idParam).Msg("Invalid user ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}
	// Parse request body
	var req struct {
		OldPassword string `json:"old_password" validate:"required"`
		NewPassword string `json:"new_password" validate:"required,min=8"`
	}

	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse change password request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if req.OldPassword == "" || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Old password and new password are required",
		})
	}

	// Change password
	err = h.userUseCase.ChangePassword(c.Context(), id, req.OldPassword, req.NewPassword)
	if err != nil {
		log.Error().Err(err).Str("id", idParam).Msg("Failed to change password")

		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		case errors.Is(err, usecase.ErrInvalidCredentials):
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid old password",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to change password",
			})
		}
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Password changed successfully",
	})
}

// UpdateStatus updates a user's status
func (h *UserHandler) UpdateStatus(c *fiber.Ctx) error {
	// Parse user ID from path
	idParam := c.Params("id")
	if idParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	// Parse UUID
	id, err := uuid.Parse(idParam)
	if err != nil {
		log.Error().Err(err).Str("id", idParam).Msg("Invalid user ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID format",
		})
	}

	// Parse request body
	var req struct {
		Status string `json:"status" validate:"required,oneof=active inactive blocked"`
	}

	if err := c.BodyParser(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse update status request body")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate status
	if req.Status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Status is required",
		})
	}

	// Update status
	err = h.userUseCase.UpdateStatus(c.Context(), id, req.Status)
	if err != nil {
		log.Error().Err(err).Str("id", idParam).Str("status", req.Status).Msg("Failed to update status")

		if errors.Is(err, usecase.ErrUserNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update status",
		})
	}

	// Return success response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Status updated successfully",
	})
}

// HealthCheck is a simple health check endpoint
func (h *UserHandler) HealthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	})
}
