package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go-fiber-template/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// RequestLoggingMiddleware creates a request logging middleware
func RequestLoggingMiddleware(cfg *config.Config) fiber.Handler {
	// Configure different log formats based on environment
	var logFormat string

	if cfg.IsDevelopment() {
		// Detailed format for development
		logFormat = "${time} | ${status} | ${latency} | ${ip} | ${method} ${path} | ${error}\n"
	} else {
		// JSON format for production
		logFormat = `{"time":"${time}","status":"${status}","latency":"${latency}","ip":"${ip}","method":"${method}","path":"${path}","user_agent":"${ua}","error":"${error}"}` + "\n"
	}

	return logger.New(logger.Config{
		Format:     logFormat,
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "UTC",
		Done: func(c *fiber.Ctx, logString []byte) {
			// Custom processing of log entries if needed
			// For now, just use the default behavior
		},
	})
}

// CustomRequestLoggingMiddleware creates a custom request logging middleware with more control
func CustomRequestLoggingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get user information if available
		userID := "anonymous"
		if user := GetUserFromContext(c); user != nil {
			userID = fmt.Sprintf("user_%d", user.GetID())
		}

		// Log request details
		log.Printf(
			"[%s] %s %s - Status: %d - Latency: %v - IP: %s - User: %s - UA: %s",
			time.Now().Format("2006-01-02 15:04:05"),
			c.Method(),
			c.Path(),
			c.Response().StatusCode(),
			latency,
			c.IP(),
			userID,
			c.Get("User-Agent"),
		)

		// Log error if present
		if err != nil {
			log.Printf("Request error: %v", err)
		}

		return err
	}
}

// StructuredLoggingMiddleware creates a structured logging middleware that outputs JSON logs
func StructuredLoggingMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get user information if available
		var userID interface{} = nil
		if user := GetUserFromContext(c); user != nil {
			userID = user.GetID()
		}

		// Create structured log entry
		logEntry := map[string]interface{}{
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
			"method":      c.Method(),
			"path":        c.Path(),
			"status_code": c.Response().StatusCode(),
			"latency_ms":  latency.Milliseconds(),
			"ip":          c.IP(),
			"user_agent":  c.Get("User-Agent"),
			"user_id":     userID,
		}

		// Add error information if present
		if err != nil {
			logEntry["error"] = err.Error()
		}

		// Add request size
		if c.Request().Header.ContentLength() > 0 {
			logEntry["request_size"] = c.Request().Header.ContentLength()
		}

		// Add response size
		logEntry["response_size"] = len(c.Response().Body())

		// Output as JSON
		jsonBytes, err := json.Marshal(logEntry)
		if err != nil {
			log.Printf("REQUEST_LOG: %+v", logEntry)
		} else {
			log.Printf("REQUEST_LOG: %s", string(jsonBytes))
		}

		return err
	}
}
