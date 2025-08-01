package routes

import (
	"go-fiber-template/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// SetupWebRoutes configures web interface routes for future web UI
func SetupWebRoutes(app *fiber.App, deps *Dependencies) {
	// Web routes group
	web := app.Group("/web")

	// Web dashboard (placeholder for future implementation)
	web.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Web interface placeholder",
			"note":    "This endpoint is reserved for future web UI implementation",
			"available_routes": fiber.Map{
				"login":     "/web/login",
				"dashboard": "/web/dashboard",
				"profile":   "/web/profile",
			},
		})
	})

	// Web authentication routes (placeholder)
	web.Get("/login", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":     "Web login page placeholder",
			"note":        "This endpoint is reserved for future web login page",
			"form_action": "/api/v1/auth/login",
		})
	})

	web.Get("/register", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":     "Web register page placeholder",
			"note":        "This endpoint is reserved for future web registration page",
			"form_action": "/api/v1/auth/register",
		})
	})

	// Web dashboard (placeholder)
	web.Get("/dashboard", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":   "Web dashboard placeholder",
			"note":      "This endpoint is reserved for future web dashboard",
			"timestamp": utils.GetCurrentTimestamp(),
		})
	})

	// Web profile page (placeholder)
	web.Get("/profile", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":      "Web profile page placeholder",
			"note":         "This endpoint is reserved for future web profile page",
			"api_endpoint": "/api/v1/users/profile",
		})
	})

	// Web user management (placeholder)
	web.Get("/users", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":      "Web users page placeholder",
			"note":         "This endpoint is reserved for future web user management page",
			"api_endpoint": "/api/v1/users",
		})
	})

	// Web settings page (placeholder)
	web.Get("/settings", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Web settings page placeholder",
			"note":    "This endpoint is reserved for future web settings page",
		})
	})
}
