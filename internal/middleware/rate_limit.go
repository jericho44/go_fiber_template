package middleware

import (
	"fmt"
	"time"

	"go-fiber-template/internal/config"
	"go-fiber-template/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimitMiddleware creates a rate limiting middleware with configurable limits
func RateLimitMiddleware(cfg *config.Config) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        cfg.Rate.Max,
		Expiration: cfg.Rate.Window,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use IP address as the key for rate limiting
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return utils.TooManyRequestsResponse(c, "Rate limit exceeded. Please try again later.")
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		Storage:                nil, // Use in-memory storage by default
	})
}

// AuthenticatedUserRateLimitMiddleware creates a rate limiting middleware for authenticated users
// This provides higher limits for authenticated users
func AuthenticatedUserRateLimitMiddleware(cfg *config.Config) fiber.Handler {
	// Authenticated users get 5x the normal rate limit
	authenticatedMax := cfg.Rate.Max * 5

	return limiter.New(limiter.Config{
		Max:        authenticatedMax,
		Expiration: cfg.Rate.Window,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Use user ID if authenticated, otherwise fall back to IP
			if user := GetUserFromContext(c); user != nil {
				return fmt.Sprintf("user_%d", user.GetID())
			}
			return "ip_" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return utils.TooManyRequestsResponse(c, "Rate limit exceeded. Please try again later.")
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		Storage:                nil, // Use in-memory storage by default
	})
}

// CustomRateLimitMiddleware creates a rate limiting middleware with custom configuration
func CustomRateLimitMiddleware(max int, window time.Duration, keyGen func(*fiber.Ctx) string) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:          max,
		Expiration:   window,
		KeyGenerator: keyGen,
		LimitReached: func(c *fiber.Ctx) error {
			return utils.TooManyRequestsResponse(c, "Rate limit exceeded. Please try again later.")
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		Storage:                nil,
	})
}

// APIEndpointRateLimitMiddleware creates different rate limits for different API endpoints
func APIEndpointRateLimitMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()
		var max int
		var window time.Duration

		// Different limits for different endpoints
		switch {
		case path == "/api/v1/auth/login" || path == "/api/v1/auth/register":
			// Stricter limits for auth endpoints
			max = cfg.Rate.Max / 10      // 10x stricter
			window = cfg.Rate.Window * 5 // 5x longer window
		case path == "/api/v1/auth/refresh":
			// Moderate limits for token refresh
			max = cfg.Rate.Max / 5       // 5x stricter
			window = cfg.Rate.Window * 2 // 2x longer window
		default:
			// Default limits for other endpoints
			max = cfg.Rate.Max
			window = cfg.Rate.Window
		}

		// Create a limiter with the determined limits
		rateLimiter := limiter.New(limiter.Config{
			Max:        max,
			Expiration: window,
			KeyGenerator: func(c *fiber.Ctx) string {
				// Use user ID if authenticated, otherwise use IP + endpoint
				if user := GetUserFromContext(c); user != nil {
					return fmt.Sprintf("user_%d_%s", user.GetID(), path)
				}
				return "ip_" + c.IP() + "_" + path
			},
			LimitReached: func(c *fiber.Ctx) error {
				return utils.TooManyRequestsResponse(c, "Rate limit exceeded for this endpoint. Please try again later.")
			},
			SkipFailedRequests:     false,
			SkipSuccessfulRequests: false,
			Storage:                nil,
		})

		return rateLimiter(c)
	}
}

// GlobalRateLimitMiddleware creates a global rate limit that applies to all requests
func GlobalRateLimitMiddleware(max int, window time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "global"
		},
		LimitReached: func(c *fiber.Ctx) error {
			return utils.TooManyRequestsResponse(c, "Global rate limit exceeded. Please try again later.")
		},
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		Storage:                nil,
	})
}
