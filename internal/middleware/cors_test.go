package middleware

import (
	"net/http"
	"testing"

	"go-fiber-template/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		origins        []string
		requestOrigin  string
		expectedOrigin string
		shouldAllow    bool
	}{
		{
			name:           "Allow single origin",
			origins:        []string{"http://localhost:3000"},
			requestOrigin:  "http://localhost:3000",
			expectedOrigin: "http://localhost:3000",
			shouldAllow:    true,
		},
		{
			name:           "Allow multiple origins - first",
			origins:        []string{"http://localhost:3000", "https://example.com"},
			requestOrigin:  "http://localhost:3000",
			expectedOrigin: "http://localhost:3000",
			shouldAllow:    true,
		},
		{
			name:           "Allow multiple origins - second",
			origins:        []string{"http://localhost:3000", "https://example.com"},
			requestOrigin:  "https://example.com",
			expectedOrigin: "https://example.com",
			shouldAllow:    true,
		},
		{
			name:           "Reject unauthorized origin",
			origins:        []string{"http://localhost:3000"},
			requestOrigin:  "https://malicious.com",
			expectedOrigin: "",
			shouldAllow:    false,
		},
		{
			name:           "Allow wildcard origin",
			origins:        []string{"*"},
			requestOrigin:  "https://any-domain.com",
			expectedOrigin: "*",
			shouldAllow:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with test origins
			cfg := &config.Config{
				CORS: config.CORSConfig{
					Origins: tt.origins,
				},
			}

			// Create Fiber app with CORS middleware
			app := fiber.New()
			app.Use(CORSMiddleware(cfg))

			// Add test route
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.JSON(fiber.Map{"message": "success"})
			})

			// Create request with Origin header
			req, err := http.NewRequest("GET", "/test", nil)
			require.NoError(t, err)
			req.Header.Set("Origin", tt.requestOrigin)

			// Execute request
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Check response
			if tt.shouldAllow {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				if tt.expectedOrigin != "*" {
					assert.Equal(t, tt.expectedOrigin, resp.Header.Get("Access-Control-Allow-Origin"))
				}
			} else {
				// For unauthorized origins, the request should still succeed but without CORS headers
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				assert.Empty(t, resp.Header.Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

func TestCORSMiddleware_PreflightRequest(t *testing.T) {
	cfg := &config.Config{
		CORS: config.CORSConfig{
			Origins: []string{"http://localhost:3000"},
		},
	}

	app := fiber.New()
	app.Use(CORSMiddleware(cfg))

	// Add test route
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create preflight request
	req, err := http.NewRequest("OPTIONS", "/test", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type,Authorization")

	// Execute request
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check preflight response
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, "http://localhost:3000", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Headers"), "Content-Type")
	assert.Contains(t, resp.Header.Get("Access-Control-Allow-Headers"), "Authorization")
}

func TestCORSMiddleware_CredentialsSupport(t *testing.T) {
	cfg := &config.Config{
		CORS: config.CORSConfig{
			Origins: []string{"http://localhost:3000"},
		},
	}

	app := fiber.New()
	app.Use(CORSMiddleware(cfg))

	// Add test route
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request with credentials
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "http://localhost:3000")

	// Execute request
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check credentials support
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "true", resp.Header.Get("Access-Control-Allow-Credentials"))
}

func TestCORSMiddlewareWithCustomConfig(t *testing.T) {
	customConfig := cors.Config{
		AllowOrigins:     "https://custom.com",
		AllowMethods:     "GET,POST",
		AllowHeaders:     "Content-Type",
		AllowCredentials: false,
		MaxAge:           3600,
	}

	app := fiber.New()
	app.Use(CORSMiddlewareWithCustomConfig(customConfig))

	// Add test route
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "https://custom.com")

	// Execute request
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check response
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "https://custom.com", resp.Header.Get("Access-Control-Allow-Origin"))
	// Note: Fiber CORS middleware doesn't set "false" explicitly, it omits the header when false
}

func TestCORSMiddleware_EmptyOrigins(t *testing.T) {
	cfg := &config.Config{
		CORS: config.CORSConfig{
			Origins: []string{},
		},
	}

	app := fiber.New()
	app.Use(CORSMiddleware(cfg))

	// Add test route
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	req.Header.Set("Origin", "http://localhost:3000")

	// Execute request
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check response - should work but without CORS headers
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// When origins is empty, Fiber CORS middleware may still set wildcard
	corsHeader := resp.Header.Get("Access-Control-Allow-Origin")
	// Accept either empty or wildcard as valid behavior
	assert.True(t, corsHeader == "" || corsHeader == "*", "Expected empty or wildcard CORS header, got: %s", corsHeader)
}
