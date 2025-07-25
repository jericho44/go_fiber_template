package utils

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPaginationParams(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected PaginationParams
	}{
		{
			name:     "default values when no query params",
			query:    "",
			expected: PaginationParams{Page: 1, Limit: 10},
		},
		{
			name:     "valid page and limit",
			query:    "?page=2&limit=20",
			expected: PaginationParams{Page: 2, Limit: 20},
		},
		{
			name:     "invalid page defaults to 1",
			query:    "?page=0&limit=15",
			expected: PaginationParams{Page: 1, Limit: 15},
		},
		{
			name:     "invalid limit defaults to 10",
			query:    "?page=3&limit=0",
			expected: PaginationParams{Page: 3, Limit: 10},
		},
		{
			name:     "limit over 100 defaults to 10",
			query:    "?page=1&limit=150",
			expected: PaginationParams{Page: 1, Limit: 10},
		},
		{
			name:     "non-numeric values default",
			query:    "?page=abc&limit=xyz",
			expected: PaginationParams{Page: 1, Limit: 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setupTestApp()

			app.Get("/test", func(c *fiber.Ctx) error {
				params := GetPaginationParams(c)
				return c.JSON(params)
			})

			req := httptest.NewRequest("GET", "/test"+tt.query, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var result PaginationParams
			err = json.Unmarshal(body, &result)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPaginationParams_CalculateOffset(t *testing.T) {
	tests := []struct {
		name     string
		params   PaginationParams
		expected int
	}{
		{
			name:     "first page",
			params:   PaginationParams{Page: 1, Limit: 10},
			expected: 0,
		},
		{
			name:     "second page",
			params:   PaginationParams{Page: 2, Limit: 10},
			expected: 10,
		},
		{
			name:     "third page with different limit",
			params:   PaginationParams{Page: 3, Limit: 20},
			expected: 40,
		},
		{
			name:     "large page number",
			params:   PaginationParams{Page: 10, Limit: 5},
			expected: 45,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.params.CalculateOffset()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateMeta(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		limit    int
		total    int64
		expected *Meta
	}{
		{
			name:  "exact division",
			page:  1,
			limit: 10,
			total: 100,
			expected: &Meta{
				Page:       1,
				Limit:      10,
				Total:      100,
				TotalPages: 10,
			},
		},
		{
			name:  "with remainder",
			page:  2,
			limit: 10,
			total: 95,
			expected: &Meta{
				Page:       2,
				Limit:      10,
				Total:      95,
				TotalPages: 10,
			},
		},
		{
			name:  "single page",
			page:  1,
			limit: 20,
			total: 15,
			expected: &Meta{
				Page:       1,
				Limit:      20,
				Total:      15,
				TotalPages: 1,
			},
		},
		{
			name:  "empty result",
			page:  1,
			limit: 10,
			total: 0,
			expected: &Meta{
				Page:       1,
				Limit:      10,
				Total:      0,
				TotalPages: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateMeta(tt.page, tt.limit, tt.total)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPaginatedResponse(t *testing.T) {
	app := setupTestApp()

	data := []map[string]string{
		{"id": "1", "name": "Item 1"},
		{"id": "2", "name": "Item 2"},
	}

	app.Get("/test", func(c *fiber.Ctx) error {
		return PaginatedResponse(c, "Items retrieved", data, 1, 10, 25)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response APIResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, "Items retrieved", response.Message)
	assert.NotNil(t, response.Data)
	assert.Nil(t, response.Error)
	assert.NotNil(t, response.Meta)
	assert.Equal(t, 1, response.Meta.Page)
	assert.Equal(t, 10, response.Meta.Limit)
	assert.Equal(t, int64(25), response.Meta.Total)
	assert.Equal(t, 3, response.Meta.TotalPages)
}
