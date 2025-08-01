package routes

import (
	"go-fiber-template/internal/config"
	"go-fiber-template/internal/controllers"
	"go-fiber-template/internal/middleware"
	"go-fiber-template/internal/repositories"
	"go-fiber-template/internal/services"
	"go-fiber-template/internal/utils"

	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// Dependencies holds all application dependencies needed for routing
type Dependencies struct {
	// Repositories
	UserRepo           repositories.UserRepository
	TokenBlacklistRepo repositories.TokenBlacklistRepository
	TxManager          repositories.TransactionManager

	// Services
	AuthService services.AuthService
	UserService services.UserService
	JWTService  services.JWTServiceInterface

	// Utils
	JWTManager *utils.JWTManager

	// Middleware
	AuthMiddleware *middleware.AuthMiddleware

	// Controllers
	AuthController   *controllers.AuthController
	UserController   *controllers.UserController
	HealthController *controllers.HealthController
	AdminController  *controllers.AdminController
}

// SetupAllRoutes configures all application routes
func SetupAllRoutes(app *fiber.App, cfg *config.Config, deps *Dependencies) {
	// Setup public routes (no authentication required)
	setupPublicRoutes(app, cfg, deps)

	// Setup API routes
	SetupAPIRoutes(app, deps)

	// Setup web routes (if needed for future web interface)
	SetupWebRoutes(app, deps)

	// Setup mobile API routes (if needed for mobile-specific endpoints)
	SetupMobileAPIRoutes(app, deps)

	// Setup documentation routes
	setupDocumentationRoutes(app)

	// Setup 404 handler (must be last)
	setup404Handler(app)
}

// setupPublicRoutes configures public routes that don't require authentication
func setupPublicRoutes(app *fiber.App, cfg *config.Config, deps *Dependencies) {
	// Root endpoint - Welcome message with available endpoints
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to Go Fiber Template API",
			"version": "1.0.0",
			"status":  "running",
			"endpoints": fiber.Map{
				"health":        "/health",
				"hello":         "/hello",
				"info":          "/info",
				"ping":          "/ping",
				"documentation": "/swagger/",
				"api_v1":        "/api/v1",
			},
		})
	})

	// Public route: Hello World - Simple greeting endpoint
	app.Get("/hello", func(c *fiber.Ctx) error {
		name := c.Query("name", "World")
		lang := c.Query("lang", "en")

		greetings := map[string]string{
			"en": "Hello",
			"es": "Hola",
			"fr": "Bonjour",
			"de": "Hallo",
			"it": "Ciao",
			"pt": "Olá",
			"ru": "Привет",
			"ja": "こんにちは",
			"ko": "안녕하세요",
			"zh": "你好",
		}

		greeting, exists := greetings[lang]
		if !exists {
			greeting = greetings["en"] // Default to English
		}

		return c.JSON(fiber.Map{
			"message":   greeting + ", " + name + "!",
			"language":  lang,
			"timestamp": utils.GetCurrentTimestamp(),
			"service":   "go-fiber-template",
			"tip":       "Try adding ?name=YourName&lang=es to customize the greeting",
		})
	})

	// Health check endpoint - Service health status
	app.Get("/health", deps.HealthController.CheckHealth)

	// API information endpoint - Service information and available endpoints
	app.Get("/info", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service":     "go-fiber-template",
			"version":     "1.0.0",
			"environment": cfg.Env,
			"description": "A comprehensive Go web application template using Fiber framework",
			"features": []string{
				"JWT Authentication",
				"User Management",
				"Database Integration",
				"API Documentation",
				"Rate Limiting",
				"CORS Support",
				"Error Handling",
				"Request Validation",
			},
			"endpoints": fiber.Map{
				"public": fiber.Map{
					"root":          "/",
					"health":        "/health",
					"hello":         "/hello",
					"info":          "/info",
					"ping":          "/ping",
					"documentation": "/swagger/",
				},
				"api_v1": fiber.Map{
					"base":  "/api/v1",
					"auth":  "/api/v1/auth",
					"users": "/api/v1/users",
				},
			},
		})
	})

	// Ping endpoint - Simple connectivity test
	app.Get("/ping", deps.HealthController.CheckHealthSimple)
}

// setupDocumentationRoutes configures API documentation routes
func setupDocumentationRoutes(app *fiber.App) {
	// Swagger documentation
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Redirect /swagger to /swagger/
	app.Get("/swagger", func(c *fiber.Ctx) error {
		return c.Redirect("/swagger/")
	})

	// Documentation info endpoint
	app.Get("/docs", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message":     "API Documentation",
			"swagger_ui":  "/swagger/",
			"description": "Interactive API documentation using Swagger UI",
		})
	})
}

// setup404Handler configures the 404 not found handler
func setup404Handler(app *fiber.App) {
	app.Use(func(c *fiber.Ctx) error {
		return utils.NotFoundResponse(c, "Route not found")
	})
}
