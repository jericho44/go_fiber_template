package utils

import (
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// PaginationParams represents pagination parameters from request
type PaginationParams struct {
	Page  int
	Limit int
}

// GetPaginationParams extracts pagination parameters from query string
func GetPaginationParams(c *fiber.Ctx) PaginationParams {
	page := 1
	limit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	return PaginationParams{
		Page:  page,
		Limit: limit,
	}
}

// CalculateOffset calculates the database offset based on page and limit
func (p PaginationParams) CalculateOffset() int {
	return (p.Page - 1) * p.Limit
}

// CreateMeta creates pagination metadata
func CreateMeta(page, limit int, total int64) *Meta {
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

// PaginatedResponse creates a paginated response with metadata
func PaginatedResponse(c *fiber.Ctx, message string, data interface{}, page, limit int, total int64) error {
	meta := CreateMeta(page, limit, total)
	return SuccessResponseWithMeta(c, message, data, meta)
}
