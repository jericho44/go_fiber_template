package routes

import (
	"go-fiber-template/internal/config"

	"github.com/gofiber/fiber/v2"
)

// SetupAPIRoutes configures all API routes for web/desktop applications
func SetupAPIRoutes(app *fiber.App, cfg *config.Config, deps *Dependencies) {
	// API v1 group
	v1 := app.Group("/api/v1")

	// Add API info endpoint for v1
	v1.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Go Fiber Template API v1",
			"version": "1.0.0",
			"platform": "web/desktop",
			"endpoints": fiber.Map{
				"auth":  "/api/v1/auth",
				"users": "/api/v1/users",
			},
		})
	})

	// Authentication routes
	SetupAuthAPIRoutes(v1, cfg, deps)

	// User management routes
	SetupUserAPIRoutes(v1, cfg, deps)
}

// SetupAuthAPIRoutes configures authentication-related routes for API
func SetupAuthAPIRoutes(api fiber.Router, cfg *config.Config, deps *Dependencies) {
	// Create auth group
	auth := api.Group("/auth")

	// Public authentication routes (no middleware required)
	auth.Post("/register", deps.AuthController.Register)
	auth.Post("/login", deps.AuthController.Login)
	auth.Post("/refresh", deps.AuthController.RefreshToken)

	// Protected authentication routes (require valid JWT token)
	authProtected := auth.Group("")
	authProtected.Use(deps.AuthMiddleware.RequireAuth())
	authProtected.Post("/logout", deps.AuthController.Logout)
}

// SetupUserAPIRoutes configures user management routes for API
func SetupUserAPIRoutes(api fiber.Router, cfg *config.Config, deps *Dependencies) {
	// Create users group with authentication middleware
	users := api.Group("/users")
	users.Use(deps.AuthMiddleware.RequireAuth())

	// User management routes
	users.Get("/", deps.UserController.ListUsers)
	users.Get("/:id", deps.UserController.GetUserProfile)
	users.Put("/:id", deps.UserController.UpdateUserProfile)
	users.Delete("/:id", deps.UserController.DeleteUser)
}