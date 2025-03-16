package router

import (
	"github.com/chats/go-user-api/api/http/handler"
	"github.com/chats/go-user-api/api/http/middleware"
	"github.com/chats/go-user-api/config"
	"github.com/chats/go-user-api/internal/domain/repository"
	"github.com/chats/go-user-api/internal/domain/service"
	"github.com/chats/go-user-api/internal/domain/usecase"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// SetupHandlers initializes all handlers and routes
func SetupHandlers(
	cfg *config.Config,
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
) (*handler.UserHandler, *handler.AuthHandler, fiber.Handler) {
	// Create token service
	tokenService, err := service.NewTokenService(cfg.Security)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create token service")
	}

	// Create use cases
	userUseCase := usecase.NewUserUseCase(userRepo)
	authUseCase := usecase.NewAuthUseCase(userRepo, tokenRepo, tokenService)

	// Create handlers
	userHandler := handler.NewUserHandler(userUseCase)
	authHandler := handler.NewAuthHandler(authUseCase)

	// Create auth middleware
	authMiddleware := middleware.AuthMiddleware(authUseCase)

	return userHandler, authHandler, authMiddleware
}
