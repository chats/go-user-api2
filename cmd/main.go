package main

import (
	"os"

	"github.com/chats/go-user-api/config"
	"github.com/chats/go-user-api/internal/logger"
	"github.com/chats/go-user-api/server"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize logger
	logger.InitLogger()

	// Load configuration
	cfg := config.LoadConfig()

	log.Info().Msg("Starting service...")

	// Create and set up server
	s := server.NewServer(cfg)
	if err := s.Setup(); err != nil {
		log.Fatal().Err(err).Msg("Failed to set up server")
		os.Exit(1)
	}

	// Start server
	if err := s.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
		os.Exit(1)
	}

}
