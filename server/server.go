package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chats/go-user-api/api/http/handler"
	"github.com/chats/go-user-api/api/http/router"
	"github.com/chats/go-user-api/config"
	"github.com/chats/go-user-api/internal/domain/repository"
	"github.com/chats/go-user-api/internal/domain/usecase"
	"github.com/chats/go-user-api/internal/infrastructure/cache"
	"github.com/chats/go-user-api/internal/infrastructure/db"

	//"github.com/chats/go-user-api/internal/infrastructure/grpc"
	//	"github.com/chats/go-user-api/internal/infrastructure/tracing"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

// Server represents the application server
type Server struct {
	config     *config.Config
	httpServer *fiber.App
	//	grpcServer     *grpc.Server
	database    db.Database
	cacheClient cache.Cache
	// tracerProvider *sdktrace.TracerProvider
}

// NewServer creates a new application server
func NewServer(cfg *config.Config) *Server {
	return &Server{
		config: cfg,
	}
}

// Setup initializes the server
func (s *Server) Setup() error {
	// Set up database
	dbFactory := db.NewDatabaseFactory()
	database, err := dbFactory.Create(s.config.Database)
	if err != nil {
		return fmt.Errorf("failed to create database: %v", err)
	}
	s.database = database

	// Connect to database
	if err := s.database.Connect(context.Background()); err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Set up cache
	cacheFactory := cache.NewCacheFactory()
	cacheClient, err := cacheFactory.Create(s.config.Cache)
	if err != nil {
		return fmt.Errorf("failed to create cache: %v", err)
	}
	s.cacheClient = cacheClient

	// Connect to cache
	if err := s.cacheClient.Connect(context.Background()); err != nil {
		return fmt.Errorf("failed to connect to cache: %v", err)
	}

	// Set up repositories
	userRepo := repository.NewUserRepository(s.database, s.cacheClient)

	// Set up use cases
	userUseCase := usecase.NewUserUseCase(userRepo)

	// Set up HTTP handlers
	userHandler := handler.NewUserHandler(userUseCase)

	// Set up gRPC server
	/*
		grpcServer, err := grpc.NewServer(grpc.Config{
			Port:             s.config.GRPC.Port,
			UseTLS:           s.config.GRPC.UseTLS,
			CertFile:         s.config.GRPC.CertFile,
			KeyFile:          s.config.GRPC.KeyFile,
			MaxRecvMsgSize:   s.config.GRPC.MaxRecvMsgSize,
			MaxSendMsgSize:   s.config.GRPC.MaxSendMsgSize,
			EnableReflection: s.config.GRPC.EnableReflection,
		})

		if err != nil {
			return fmt.Errorf("failed to create gRPC server: %v", err)
		}

		// Register gRPC service
		userService := grpcService.NewUserService(userUseCase)
		grpcServer.RegisterUserService(userService.(proto.UserServiceServer))
		s.grpcServer = grpcServer
	*/

	// Set up HTTP server
	//routes.SetupRoutes(app, cfg, authHandler, userHandler, roleHandler, permissionHandler, authService)

	httpServer := router.Setup(s.config, userHandler)
	s.httpServer = httpServer

	return nil
}

// Start starts the server
func (s *Server) Start() error {
	// Start HTTP server
	go func() {
		log.Info().Int("port", s.config.HTTP.Port).Msg("Starting HTTP server")
		if err := s.httpServer.Listen(fmt.Sprintf(":%d", s.config.HTTP.Port)); err != nil {
			log.Fatal().Err(err).Msg("Failed to start HTTP server")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server")

	// Create a timeout context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := s.httpServer.ShutdownWithContext(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to shutdown HTTP server gracefully")
	}
	// Close database connection
	if err := s.database.Close(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to close database connection")
	}

	// Close cache connection
	if err := s.cacheClient.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close cache connection")
	}

	log.Info().Msg("Server gracefully stopped")
	return nil
}

// GetHTTPServer returns the HTTP server
func (s *Server) GetHTTPServer() *fiber.App {
	return s.httpServer
}

// CheckRepositories is a helper function to check if all necessary repositories are registered
func CheckRepositories(ur repository.UserRepository) bool {
	return ur != nil
}

// CheckUseCases is a helper function to check if all necessary use cases are registered
func CheckUseCases(uu usecase.UserUseCase) bool {
	return uu != nil
}
