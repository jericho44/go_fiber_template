package middleware

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-fiber-template/internal/config"
	"go-fiber-template/internal/services"
	"go-fiber-template/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// EnhancedRateLimitMiddleware creates an enhanced rate limiting middleware with Redis support
type EnhancedRateLimitMiddleware struct {
	rateLimiter services.RateLimiter
	config      *config.Config
}

// NewEnhancedRateLimitMiddleware creates a new enhanced rate limit middleware
func NewEnhancedRateLimitMiddleware(cfg *config.Config) (*EnhancedRateLimitMiddleware, error) {
	rateLimiter, err := services.NewRateLimiter(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create rate limiter: %v", err)
	}

	return &EnhancedRateLimitMiddleware{
		rateLimiter: rateLimiter,
		config:      cfg,
	}, nil
}

// PerUserAndIPRateLimit creates middleware that applies rate limiting per user and per IP
func (e *EnhancedRateLimitMiddleware) PerUserAndIPRateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()

		// Check IP-based rate limit first
		ip := c.IP()
		ipResult, err := e.rateLimiter.CheckIPLimit(ctx, ip, e.config.Rate.Max, e.config.Rate.Window)
		if err != nil {
			return utils.InternalServerErrorResponse(c, "Rate limit check failed")
		}

		// Set IP rate limit headers
		c.Set("X-RateLimit-IP-Limit", strconv.Itoa(ipResult.Limit))
		c.Set("X-RateLimit-IP-Remaining", strconv.Itoa(ipResult.Remaining))
		c.Set("X-RateLimit-IP-Reset", strconv.FormatInt(ipResult.ResetTime.Unix(), 10))

		if !ipResult.Allowed {
			c.Set("Retry-After", strconv.Itoa(int(ipResult.RetryAfter.Seconds())))
			return utils.TooManyRequestsResponse(c, "IP rate limit exceeded. Please try again later.")
		}

		// Check user-based rate limit if authenticated
		if user := GetUserFromContext(c); user != nil {
			// Authenticated users get higher limits
			userLimit := e.config.Rate.Max * 5
			userResult, err := e.rateLimiter.CheckUserLimit(ctx, user.GetID(), userLimit, e.config.Rate.Window)
			if err != nil {
				return utils.InternalServerErrorResponse(c, "User rate limit check failed")
			}

			// Set user rate limit headers
			c.Set("X-RateLimit-User-Limit", strconv.Itoa(userResult.Limit))
			c.Set("X-RateLimit-User-Remaining", strconv.Itoa(userResult.Remaining))
			c.Set("X-RateLimit-User-Reset", strconv.FormatInt(userResult.ResetTime.Unix(), 10))

			if !userResult.Allowed {
				c.Set("Retry-After", strconv.Itoa(int(userResult.RetryAfter.Seconds())))
				return utils.TooManyRequestsResponse(c, "User rate limit exceeded. Please try again later.")
			}
		}

		return c.Next()
	}
}

// EndpointSpecificRateLimit creates middleware with different limits for different endpoints
func (e *EnhancedRateLimitMiddleware) EndpointSpecificRateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		path := c.Path()

		var limit int
		var window time.Duration

		// Different limits for different endpoints
		switch {
		case strings.Contains(path, "/auth/login") || strings.Contains(path, "/auth/register"):
			// Stricter limits for auth endpoints
			limit = e.config.Rate.Max / 10
			window = e.config.Rate.Window * 5
		case strings.Contains(path, "/auth/refresh"):
			// Moderate limits for token refresh
			limit = e.config.Rate.Max / 5
			window = e.config.Rate.Window * 2
		case strings.Contains(path, "/auth/"):
			// General auth endpoints
			limit = e.config.Rate.Max / 3
			window = e.config.Rate.Window * 2
		default:
			// Default limits for other endpoints
			limit = e.config.Rate.Max
			window = e.config.Rate.Window
		}

		// Create identifier (user ID if authenticated, otherwise IP)
		identifier := c.IP()
		if user := GetUserFromContext(c); user != nil {
			identifier = fmt.Sprintf("user:%d", user.GetID())
		} else {
			identifier = fmt.Sprintf("ip:%s", identifier)
		}

		result, err := e.rateLimiter.CheckEndpointLimit(ctx, identifier, path, limit, window)
		if err != nil {
			return utils.InternalServerErrorResponse(c, "Endpoint rate limit check failed")
		}

		// Set rate limit headers
		c.Set("X-RateLimit-Endpoint-Limit", strconv.Itoa(result.Limit))
		c.Set("X-RateLimit-Endpoint-Remaining", strconv.Itoa(result.Remaining))
		c.Set("X-RateLimit-Endpoint-Reset", strconv.FormatInt(result.ResetTime.Unix(), 10))

		if !result.Allowed {
			c.Set("Retry-After", strconv.Itoa(int(result.RetryAfter.Seconds())))
			return utils.TooManyRequestsResponse(c, fmt.Sprintf("Rate limit exceeded for endpoint %s. Please try again later.", path))
		}

		return c.Next()
	}
}

// AdminBypassRateLimit creates middleware that bypasses rate limits for admin users
func (e *EnhancedRateLimitMiddleware) AdminBypassRateLimit(next fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if user is admin (you can implement admin check logic here)
		if user := GetUserFromContext(c); user != nil {
			// For now, we'll assume admin check based on user ID or email
			// In a real implementation, you'd check user roles/permissions
			if e.isAdminUser(user.GetEmail()) {
				// Skip rate limiting for admin users
				return c.Next()
			}
		}

		// Apply rate limiting for non-admin users
		return next(c)
	}
}

// isAdminUser checks if a user is an admin (placeholder implementation)
func (e *EnhancedRateLimitMiddleware) isAdminUser(email string) bool {
	// This is a placeholder implementation
	// In a real application, you would check user roles/permissions from the database
	adminEmails := []string{"admin@example.com", "superuser@example.com"}
	for _, adminEmail := range adminEmails {
		if email == adminEmail {
			return true
		}
	}
	return false
}

// Legacy middleware functions for backward compatibility

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
