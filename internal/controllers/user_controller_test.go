package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-fiber-template/internal/models"
	"go-fiber-template/internal/services"
	"go-fiber-template/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService is a mock implementation of UserService for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetUserProfile(ctx context.Context, userID uint) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateUserProfile(ctx context.Context, userID uint, req services.UpdateUserProfileRequest) (*models.User, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) ListUsers(ctx context.Context, params services.ListUsersParams) (*services.ListUsersResponse, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ListUsersResponse), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// Helper function to create a test Fiber app with user controller
func setupUserControllerTest() (*fiber.App, *MockUserService) {
	mockUserService := &MockUserService{}
	controller := NewUserController(mockUserService)

	app := fiber.New()
	app.Get("/users/:id", controller.GetUserProfile)
	app.Put("/users/:id", controller.UpdateUserProfile)
	app.Get("/users", controller.ListUsers)
	app.Delete("/users/:id", controller.DeleteUser)

	return app, mockUserService
}

func TestUserController_GetUserProfile(t *testing.T) {
	t.Run("Success - Get user profile", func(t *testing.T) {
		app, mockService := setupUserControllerTest()
		expectedUser := &models.User{
			ID:        1,
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			IsActive:  true,
		}

		mockService.On("GetUserProfile", mock.Anything, uint(1)).Return(expectedUser, nil)

		req := httptest.NewRequest("GET", "/users/1", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "User profile retrieved successfully", response.Message)

		mockService.AssertExpectations(t)
	})

	t.Run("Error - Invalid user ID", func(t *testing.T) {
		app, _ := setupUserControllerTest()
		req := httptest.NewRequest("GET", "/users/invalid", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "Invalid user ID", response.Error.Message)
	})

	t.Run("Error - User not found", func(t *testing.T) {
		app, mockService := setupUserControllerTest()
		mockService.On("GetUserProfile", mock.Anything, uint(999)).Return(nil, utils.NewNotFoundError("User not found"))

		req := httptest.NewRequest("GET", "/users/999", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "User not found", response.Error.Message)

		mockService.AssertExpectations(t)
	})

	t.Run("Error - Service error", func(t *testing.T) {
		app, mockService := setupUserControllerTest()
		mockService.On("GetUserProfile", mock.Anything, uint(1)).Return(nil, fmt.Errorf("database error"))

		req := httptest.NewRequest("GET", "/users/1", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "Failed to get user profile", response.Error.Message)

		mockService.AssertExpectations(t)
	})
}

func TestUserController_UpdateUserProfile(t *testing.T) {
	t.Run("Success - Update user profile", func(t *testing.T) {
		app, mockService := setupUserControllerTest()
		firstName := "Jane"
		lastName := "Smith"
		isActive := false

		requestBody := UpdateUserProfileRequest{
			FirstName: &firstName,
			LastName:  &lastName,
			IsActive:  &isActive,
		}

		expectedUser := &models.User{
			ID:        1,
			Email:     "test@example.com",
			FirstName: firstName,
			LastName:  lastName,
			IsActive:  isActive,
		}

		expectedServiceReq := services.UpdateUserProfileRequest{
			FirstName: &firstName,
			LastName:  &lastName,
			IsActive:  &isActive,
		}

		mockService.On("UpdateUserProfile", mock.Anything, uint(1), expectedServiceReq).Return(expectedUser, nil)

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("PUT", "/users/1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "User profile updated successfully", response.Message)

		mockService.AssertExpectations(t)
	})

	t.Run("Success - Partial update", func(t *testing.T) {
		app, mockService := setupUserControllerTest()
		firstName := "Jane"

		requestBody := UpdateUserProfileRequest{
			FirstName: &firstName,
		}

		expectedUser := &models.User{
			ID:        1,
			Email:     "test@example.com",
			FirstName: firstName,
			LastName:  "Doe",
			IsActive:  true,
		}

		expectedServiceReq := services.UpdateUserProfileRequest{
			FirstName: &firstName,
		}

		mockService.On("UpdateUserProfile", mock.Anything, uint(1), expectedServiceReq).Return(expectedUser, nil)

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("PUT", "/users/1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)

		mockService.AssertExpectations(t)
	})

	t.Run("Error - Invalid user ID", func(t *testing.T) {
		app, _ := setupUserControllerTest()
		requestBody := UpdateUserProfileRequest{}
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("PUT", "/users/invalid", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Error - Invalid request body", func(t *testing.T) {
		app, _ := setupUserControllerTest()
		req := httptest.NewRequest("PUT", "/users/1", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Error - User not found", func(t *testing.T) {
		app, mockService := setupUserControllerTest()
		requestBody := UpdateUserProfileRequest{}
		expectedServiceReq := services.UpdateUserProfileRequest{}

		mockService.On("UpdateUserProfile", mock.Anything, uint(999), expectedServiceReq).Return(nil, utils.NewNotFoundError("User not found"))

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("PUT", "/users/999", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		mockService.AssertExpectations(t)
	})
}

func TestUserController_ListUsers(t *testing.T) {
	t.Run("Success - List users with default pagination", func(t *testing.T) {
		app, mockService := setupUserControllerTest()
		expectedUsers := []models.User{
			{ID: 1, Email: "user1@example.com", FirstName: "John", LastName: "Doe", IsActive: true},
			{ID: 2, Email: "user2@example.com", FirstName: "Jane", LastName: "Smith", IsActive: true},
		}

		expectedResponse := &services.ListUsersResponse{
			Users: expectedUsers,
			Total: 2,
		}

		expectedParams := services.ListUsersParams{
			Page:  1,
			Limit: 10,
		}

		mockService.On("ListUsers", mock.Anything, expectedParams).Return(expectedResponse, nil)

		req := httptest.NewRequest("GET", "/users", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "Users retrieved successfully", response.Message)
		assert.NotNil(t, response.Meta)
		assert.Equal(t, 1, response.Meta.Page)
		assert.Equal(t, 10, response.Meta.Limit)
		assert.Equal(t, int64(2), response.Meta.Total)

		mockService.AssertExpectations(t)
	})

	t.Run("Success - List users with custom pagination and filters", func(t *testing.T) {
		app, mockService := setupUserControllerTest()
		expectedUsers := []models.User{
			{ID: 1, Email: "john@example.com", FirstName: "John", LastName: "Doe", IsActive: true},
		}

		expectedResponse := &services.ListUsersResponse{
			Users: expectedUsers,
			Total: 1,
		}

		isActive := true
		expectedParams := services.ListUsersParams{
			Page:      2,
			Limit:     5,
			Search:    "john",
			IsActive:  &isActive,
			SortBy:    "first_name",
			SortOrder: "asc",
		}

		mockService.On("ListUsers", mock.Anything, expectedParams).Return(expectedResponse, nil)

		req := httptest.NewRequest("GET", "/users?page=2&limit=5&search=john&is_active=true&sort_by=first_name&sort_order=asc", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, 2, response.Meta.Page)
		assert.Equal(t, 5, response.Meta.Limit)

		mockService.AssertExpectations(t)
	})

	t.Run("Error - Service error", func(t *testing.T) {
		app, mockService := setupUserControllerTest()
		expectedParams := services.ListUsersParams{
			Page:  1,
			Limit: 10,
		}

		mockService.On("ListUsers", mock.Anything, expectedParams).Return(nil, fmt.Errorf("database error"))

		req := httptest.NewRequest("GET", "/users", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "Failed to list users", response.Error.Message)

		mockService.AssertExpectations(t)
	})
}

func TestUserController_DeleteUser(t *testing.T) {
	t.Run("Success - Delete user", func(t *testing.T) {
		app, mockService := setupUserControllerTest()
		mockService.On("DeleteUser", mock.Anything, uint(1)).Return(nil)

		req := httptest.NewRequest("DELETE", "/users/1", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "User deleted successfully", response.Message)

		mockService.AssertExpectations(t)
	})

	t.Run("Error - Invalid user ID", func(t *testing.T) {
		app, _ := setupUserControllerTest()
		req := httptest.NewRequest("DELETE", "/users/invalid", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "Invalid user ID", response.Error.Message)
	})

	t.Run("Error - User not found", func(t *testing.T) {
		app, mockService := setupUserControllerTest()
		mockService.On("DeleteUser", mock.Anything, uint(999)).Return(utils.NewNotFoundError("User not found"))

		req := httptest.NewRequest("DELETE", "/users/999", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Error.Message, "User not found")

		mockService.AssertExpectations(t)
	})

	t.Run("Error - Service error", func(t *testing.T) {
		app, mockService := setupUserControllerTest()
		mockService.On("DeleteUser", mock.Anything, uint(1)).Return(fmt.Errorf("database error"))

		req := httptest.NewRequest("DELETE", "/users/1", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "Failed to delete user", response.Error.Message)

		mockService.AssertExpectations(t)
	})
}

func TestUserController_ValidationErrors(t *testing.T) {
	t.Run("Validation error - Invalid first name length", func(t *testing.T) {
		app, _ := setupUserControllerTest()
		emptyString := ""
		requestBody := UpdateUserProfileRequest{
			FirstName: &emptyString,
		}

		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("PUT", "/users/1", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Contains(t, response.Error.Message, "Validation failed")
	})
}

func TestUserController_EdgeCases(t *testing.T) {
	t.Run("Edge case - Zero user ID", func(t *testing.T) {
		app, mockService := setupUserControllerTest()
		expectedUser := &models.User{
			ID:        0,
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
			IsActive:  true,
		}
		mockService.On("GetUserProfile", mock.Anything, uint(0)).Return(expectedUser, nil)

		req := httptest.NewRequest("GET", "/users/0", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		// Should still be valid as 0 is a valid uint, but service should handle it
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockService.AssertExpectations(t)
	})

	t.Run("Edge case - Large user ID", func(t *testing.T) {
		app, mockService := setupUserControllerTest()
		mockService.On("GetUserProfile", mock.Anything, uint(4294967295)).Return(nil, utils.NewNotFoundError("User not found"))

		req := httptest.NewRequest("GET", "/users/4294967295", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		mockService.AssertExpectations(t)
	})

	t.Run("Edge case - Empty request body for update", func(t *testing.T) {
		app, mockService := setupUserControllerTest()
		expectedUser := &models.User{
			ID:        1,
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
			IsActive:  true,
		}

		expectedServiceReq := services.UpdateUserProfileRequest{}

		mockService.On("UpdateUserProfile", mock.Anything, uint(1), expectedServiceReq).Return(expectedUser, nil)

		req := httptest.NewRequest("PUT", "/users/1", bytes.NewReader([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		mockService.AssertExpectations(t)
	})
}
