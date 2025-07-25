package middleware

import (
	"log"
	"net/http"

	"go-fiber-template/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// ErrorHandler creates a middleware for handling errors consistently
func ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Check if it's a custom AppError
		if appErr, ok := utils.AsAppError(err); ok {
			return utils.ErrorResponse(c, appErr.StatusCode, appErr.Code, appErr.Message, appErr.Details)
		}

		// Check if it's a Fiber error
		if fiberErr, ok := err.(*fiber.Error); ok {
			code := getErrorCodeFromStatus(fiberErr.Code)
			return utils.ErrorResponse(c, fiberErr.Code, code, fiberErr.Message, nil)
		}

		// Log unexpected errors
		log.Printf("Unexpected error: %v", err)

		// Default to internal server error
		return utils.InternalServerErrorResponse(c, "An unexpected error occurred")
	}
}

// getErrorCodeFromStatus maps HTTP status codes to error codes
func getErrorCodeFromStatus(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusMethodNotAllowed:
		return "METHOD_NOT_ALLOWED"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusUnprocessableEntity:
		return "UNPROCESSABLE_ENTITY"
	case http.StatusTooManyRequests:
		return "TOO_MANY_REQUESTS"
	case http.StatusInternalServerError:
		return "INTERNAL_SERVER_ERROR"
	case http.StatusBadGateway:
		return "BAD_GATEWAY"
	case http.StatusServiceUnavailable:
		return "SERVICE_UNAVAILABLE"
	case http.StatusGatewayTimeout:
		return "GATEWAY_TIMEOUT"
	default:
		return "UNKNOWN_ERROR"
	}
}

// RecoverMiddleware creates a middleware for panic recovery
func RecoverMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic recovered: %v", r)

				// Create an internal server error
				err := utils.NewInternalServerError("Internal server error", nil)

				// Use the error handler to respond
				utils.ErrorResponse(c, err.StatusCode, err.Code, err.Message, err.Details)
			}
		}()

		return c.Next()
	}
}
