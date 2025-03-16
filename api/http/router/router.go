package router

import (
	"time"

	"github.com/chats/go-user-api/api/http/handler"
	"github.com/chats/go-user-api/config"
	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/rs/zerolog/log"
)

// Setup sets up the fiber router with middleware and routes
func Setup(
	cfg *config.Config,
	userHandler *handler.UserHandler,
	authHandler *handler.AuthHandler,
	authMiddleware fiber.Handler,
) *fiber.App {
	// Create new Fiber app
	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  2 * cfg.HTTP.IdleTimeout,
		Prefork:      cfg.HTTP.EnablePrefork,
		AppName:      cfg.App.Name,
	})

	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: &log.Logger,
	}))

	// Add request ID middleware
	if cfg.Middleware.EnableRequestID {
		app.Use(requestid.New())
	}

	// Add recover middleware
	if cfg.Middleware.EnableRecover {
		app.Use(recover.New(recover.Config{
			EnableStackTrace: true,
		}))
	}

	// Add CORS middleware
	if cfg.Middleware.EnableCORS {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     "*",
			AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
			AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Request-ID",
			ExposeHeaders:    "Content-Length, X-Request-ID",
			AllowCredentials: true,
			MaxAge:           86400, // 24 hours
		}))
	}

	// Add helmet middleware
	if cfg.Middleware.EnableHelmet {
		app.Use(helmet.New())
	}

	// Add rate limiter middleware
	if cfg.Middleware.EnableRateLimiter {
		app.Use(limiter.New(limiter.Config{
			Max:        100,
			Expiration: 1 * time.Minute,
			KeyGenerator: func(c *fiber.Ctx) string {
				return c.IP()
			},
			LimitReached: func(c *fiber.Ctx) error {
				log.Warn().Str("ip", c.IP()).Msg("Rate limit reached")
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error": "Too many requests, please try again later",
				})
			},
		}))
	}

	// Add ETag middleware
	if cfg.Middleware.EnableETag {
		app.Use(etag.New())
	}

	// Add compression middleware
	if cfg.Middleware.EnableCompression {
		app.Use(compress.New(compress.Config{
			Level: compress.LevelBestSpeed,
		}))
	}

	// Setup routes
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Register health check route
	api.Get("/health", userHandler.HealthCheck)

	// Register user/auth routes
	userHandler.RegisterRoutes(v1, authMiddleware)
	authHandler.RegisterRoutes(v1, authMiddleware)

	// 404 Handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Endpoint not found",
		})
	})

	return app
}
