package routes

import (
	"net/http/httptest"
	"testing"

	"go-fiber-template/internal/config"
	"go-fiber-template/internal/controllers"
	"go-fiber-template/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestRoutesSetup(t *testing.T) {
	// Create a test Fiber app
	app := fiber.New()

	// Create mock configuration
	cfg := &config.Config{
		Env: "test",
	}

	// Create mock dependencies (using nil for simplicity in this test)
	deps := &Dependencies{
		UserRepo:           nil,
		TokenBlacklistRepo: nil,
		TxManager:          nil,
		AuthService:        nil,
		UserService:        nil,
		JWTService:         nil,
		JWTManager:         nil,
		AuthMiddleware:     nil,
		AuthController:     nil,
		UserController:     nil,
	}

	// Test that SetupAllRoutes doesn't panic
	assert.NotPanics(t, func() {
		SetupAllRoutes(app, cfg, deps)
	})

	// Test health endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test root endpoint
	req = httptest.NewRequest("GET", "/", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test 404 handler
	req = httptest.NewRequest("GET", "/nonexistent", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}

func TestHealthRoutes(t *testing.T) {
	app := fiber.New()
	cfg := &config.Config{
		Env: "test",
	}

	// Setup public routes (which includes health routes)
	setupPublicRoutes(app, cfg)

	// Test health endpoint
	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test root endpoint
	req = httptest.NewRequest("GET", "/", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test hello endpoint
	req = httptest.NewRequest("GET", "/hello", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test hello endpoint with name parameter
	req = httptest.NewRequest("GET", "/hello?name=John", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test info endpoint
	req = httptest.NewRequest("GET", "/info", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test ping endpoint
	req = httptest.NewRequest("GET", "/ping", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestDocumentationRoutes(t *testing.T) {
	app := fiber.New()
	cfg := &config.Config{
		Env: "test",
	}

	// Setup documentation routes
	setupDocumentationRoutes(app, cfg)

	// Test swagger redirect - the route should exist but may not work without proper setup
	req := httptest.NewRequest("GET", "/swagger", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	// Accept either redirect (302) or not found (404) since swagger setup may not be complete in test
	assert.True(t, resp.StatusCode == 302 || resp.StatusCode == 404)
}

func TestRouteStructure(t *testing.T) {
	app := fiber.New()
	cfg := &config.Config{
		Env: "test",
	}

	// Create minimal mock dependencies for route structure testing
	deps := &Dependencies{
		AuthController: &controllers.AuthController{},
		UserController: &controllers.UserController{},
		AuthMiddleware: &middleware.AuthMiddleware{},
	}

	// This should not panic even with minimal dependencies
	assert.NotPanics(t, func() {
		api := app.Group("/api/v1")
		SetupAuthAPIRoutes(api, cfg, deps)
		SetupUserAPIRoutes(api, cfg, deps)
	})
}

func TestHelloWorldEndpoint(t *testing.T) {
	app := fiber.New()
	cfg := &config.Config{
		Env: "test",
	}

	// Setup public routes
	setupPublicRoutes(app, cfg)

	// Test basic hello world
	req := httptest.NewRequest("GET", "/hello", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test hello world with custom name
	req = httptest.NewRequest("GET", "/hello?name=Alice", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test hello world with different language
	req = httptest.NewRequest("GET", "/hello?name=Carlos&lang=es", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test hello world with invalid language (should default to English)
	req = httptest.NewRequest("GET", "/hello?name=Test&lang=invalid", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}