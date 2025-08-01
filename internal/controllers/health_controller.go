package controllers

import (
	"go-fiber-template/internal/services"
	"go-fiber-template/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// HealthController handles health check endpoints
// This controller provides comprehensive health monitoring
type HealthController struct {
	healthService *services.HealthService
}

// NewHealthController creates a new health controller
func NewHealthController(healthService *services.HealthService) *HealthController {
	return &HealthController{
		healthService: healthService,
	}
}

// CheckHealth handles the health check endpoint
// @Summary Check application health
// @Description Get comprehensive health status of the application including database, memory, and disk space
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} services.HealthStatus "Health status"
// @Success 503 {object} services.HealthStatus "Service unavailable"
// @Router /health [get]
func (h *HealthController) CheckHealth(c *fiber.Ctx) error {
	ctx := c.Context()

	health := h.healthService.CheckHealth(ctx)

	// Return appropriate HTTP status based on health
	statusCode := fiber.StatusOK
	if health.Status == "unhealthy" {
		statusCode = fiber.StatusServiceUnavailable
	} else if health.Status == "degraded" {
		statusCode = fiber.StatusOK // Still return 200 for degraded but functional
	}

	return c.Status(statusCode).JSON(health)
}

// CheckHealthSimple handles a simple health check endpoint
// @Summary Simple health check
// @Description Get basic health status - returns 200 if service is running
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} utils.APIResponse "Service is running"
// @Router /ping [get]
func (h *HealthController) CheckHealthSimple(c *fiber.Ctx) error {
	return utils.SuccessResponse(c, "pong", fiber.Map{
		"timestamp": utils.GetCurrentTimestamp(),
		"latency":   "< 1ms",
	})
}
