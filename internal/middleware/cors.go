package middleware

import (
	"strings"

	"go-fiber-template/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSMiddleware creates a CORS middleware with configurable origins
func CORSMiddleware(cfg *config.Config) fiber.Handler {
	// Convert origins slice to comma-separated string for Fiber CORS middleware
	allowOrigins := strings.Join(cfg.CORS.Origins, ",")

	return cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length,Content-Type",
		MaxAge:           86400, // 24 hours
	})
}

// CORSMiddlewareWithCustomConfig creates a CORS middleware with custom configuration
func CORSMiddlewareWithCustomConfig(corsConfig cors.Config) fiber.Handler {
	return cors.New(corsConfig)
}
