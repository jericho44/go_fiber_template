package routes

import (
	"go-fiber-template/internal/config"

	"github.com/gofiber/fiber/v2"
)

// SetupMobileAPIRoutes configures mobile-specific API routes
func SetupMobileAPIRoutes(app *fiber.App, cfg *config.Config, deps *Dependencies) {
	// Mobile API v1 group
	mobile := app.Group("/api/mobile/v1")

	// Mobile API info endpoint
	mobile.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Go Fiber Template Mobile API v1",
			"version": "1.0.0",
			"platform": "mobile",
			"endpoints": fiber.Map{
				"auth":  "/api/mobile/v1/auth",
				"users": "/api/mobile/v1/users",
			},
		})
	})

	// Mobile authentication routes (same as API but with mobile-specific optimizations)
	SetupAuthMobileRoutes(mobile, cfg, deps)

	// Mobile user management routes (same as API but with mobile-specific responses)
	SetupUserMobileRoutes(mobile, cfg, deps)
}

// SetupAuthMobileRoutes configures authentication routes for mobile API
func SetupAuthMobileRoutes(mobile fiber.Router, cfg *config.Config, deps *Dependencies) {
	// Create auth group for mobile
	auth := mobile.Group("/auth")

	// Public authentication routes (no middleware required)
	auth.Post("/register", deps.AuthController.Register)
	auth.Post("/login", deps.AuthController.Login)
	auth.Post("/refresh", deps.AuthController.RefreshToken)

	// Protected authentication routes (require valid JWT token)
	authProtected := auth.Group("")
	authProtected.Use(deps.AuthMiddleware.RequireAuth())
	authProtected.Post("/logout", deps.AuthController.Logout)

	// Mobile-specific auth endpoints (placeholder for future mobile features)
	authProtected.Get("/device-info", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Device info endpoint placeholder",
			"note":    "This endpoint is reserved for mobile device management",
			"features": []string{
				"Device registration",
				"Push notification tokens",
				"Device-specific settings",
				"Biometric authentication",
			},
		})
	})

	// Mobile-specific push notification management
	authProtected.Post("/push-token", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Push token registration placeholder",
			"note":    "This endpoint is reserved for mobile push notification token management",
		})
	})

	// Mobile-specific biometric authentication
	authProtected.Post("/biometric", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Biometric authentication placeholder",
			"note":    "This endpoint is reserved for mobile biometric authentication",
		})
	})
}

// SetupUserMobileRoutes configures user management routes for mobile API
func SetupUserMobileRoutes(mobile fiber.Router, cfg *config.Config, deps *Dependencies) {
	// Create users group with authentication middleware
	users := mobile.Group("/users")
	users.Use(deps.AuthMiddleware.RequireAuth())

	// User management routes (same as API)
	users.Get("/", deps.UserController.ListUsers)
	users.Get("/:id", deps.UserController.GetUserProfile)
	users.Put("/:id", deps.UserController.UpdateUserProfile)
	users.Delete("/:id", deps.UserController.DeleteUser)

	// Mobile-specific user endpoints (placeholder for future mobile features)
	users.Get("/profile/mobile", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Mobile profile endpoint placeholder",
			"note":    "This endpoint is reserved for mobile-optimized profile data",
			"features": []string{
				"Compressed profile data",
				"Mobile-specific fields",
				"Offline sync support",
			},
		})
	})

	// Mobile-specific settings
	users.Get("/settings/mobile", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Mobile settings endpoint placeholder",
			"note":    "This endpoint is reserved for mobile-specific user settings",
			"settings": []string{
				"Push notifications",
				"Offline mode",
				"Data usage preferences",
				"Biometric settings",
			},
		})
	})

	// Mobile-specific avatar upload (optimized for mobile)
	users.Post("/avatar/mobile", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Mobile avatar upload placeholder",
			"note":    "This endpoint is reserved for mobile-optimized avatar upload",
			"features": []string{
				"Image compression",
				"Multiple format support",
				"Offline queue support",
			},
		})
	})
}