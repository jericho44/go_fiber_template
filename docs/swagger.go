// Package docs contains the Swagger documentation for the Go Fiber Template API
//
// @title Go Fiber Template API
// @version 1.0
// @description A comprehensive Go web application template using the Fiber framework with essential features including API documentation, database management, structured response handling, clean architecture patterns, and secure authentication.
// @description
// @description ## 🚀 Quick Start Guide
// @description
// @description ### 1. Authentication Flow
// @description This API uses JWT (JSON Web Token) for authentication. Follow these steps to get started:
// @description
// @description **Step 1: Register or Login**
// @description - **Register**: Use `POST /auth/register` to create a new account
// @description - **Login**: Use `POST /auth/login` with existing credentials
// @description
// @description **Step 2: Authorize in Swagger**
// @description 1. Copy the `access_token` from the login/register response
// @description 2. Click the **"Authorize"** button at the top of this page
// @description 3. Enter: `Bearer {your-access-token}` (replace with your actual token)
// @description 4. Click **"Authorize"** to save the token
// @description
// @description **Step 3: Test Protected Endpoints**
// @description - All User Management endpoints now include your authorization automatically
// @description - Try the `GET /users` endpoint to list users
// @description
// @description **Step 4: Token Management**
// @description - **Access tokens expire in 15 minutes** - use `POST /auth/refresh` to get new tokens
// @description - **Logout** using `POST /auth/logout` to invalidate tokens when done
// @description
// @description ### 2. Example Authentication Flow
// @description ```
// @description 1. POST /auth/register → Get access_token and refresh_token
// @description 2. Click "Authorize" → Enter "Bearer {access_token}"
// @description 3. Test protected endpoints (GET /users, PUT /users/{id}, etc.)
// @description 4. When token expires → POST /auth/refresh with refresh_token
// @description 5. POST /auth/logout → Clean logout and token invalidation
// @description ```
// @description
// @description ## 📋 API Response Format
// @description All API responses follow a consistent structure for predictable client integration:
// @description
// @description **Success Response:**
// @description ```json
// @description {
// @description   "success": true,
// @description   "message": "Operation successful",
// @description   "data": { ... },
// @description   "error": null,
// @description   "meta": { ... }
// @description }
// @description ```
// @description
// @description **Error Response:**
// @description ```json
// @description {
// @description   "success": false,
// @description   "message": "Request failed",
// @description   "data": null,
// @description   "error": {
// @description     "code": "VALIDATION_ERROR",
// @description     "message": "Validation failed",
// @description     "details": { "email": "Email is required" }
// @description   },
// @description   "meta": null
// @description }
// @description ```
// @description
// @description ## ⚠️ Error Handling
// @description The API uses standard HTTP status codes and provides detailed error information:
// @description
// @description **Common Error Codes:**
// @description - `VALIDATION_ERROR` (400): Request validation failed
// @description - `UNAUTHORIZED` (401): Invalid credentials or expired token
// @description - `FORBIDDEN` (403): Access denied
// @description - `NOT_FOUND` (404): Resource not found
// @description - `CONFLICT` (409): Resource already exists (e.g., email already registered)
// @description - `INTERNAL_ERROR` (500): Server error
// @description
// @description **Error Response Fields:**
// @description - `success`: Always `false` for errors
// @description - `message`: Always "Request failed" for errors
// @description - `error.code`: Specific error code for programmatic handling
// @description - `error.message`: Human-readable error message
// @description - `error.details`: Field-specific validation errors (when applicable)
// @description
// @description ## 📄 Pagination
// @description List endpoints support pagination with consistent query parameters:
// @description
// @description **Query Parameters:**
// @description - `page`: Page number (default: 1, minimum: 1)
// @description - `limit`: Items per page (default: 10, minimum: 1, maximum: 100)
// @description
// @description **Pagination Response:**
// @description ```json
// @description {
// @description   "success": true,
// @description   "message": "Users retrieved successfully",
// @description   "data": [ ... ],
// @description   "meta": {
// @description     "page": 1,
// @description     "limit": 10,
// @description     "total": 100,
// @description     "total_pages": 10
// @description   }
// @description }
// @description ```
// @description
// @description ## 🔍 Search and Filtering
// @description User endpoints support advanced search and filtering:
// @description
// @description **Search Parameters:**
// @description - `search`: Search in first_name, last_name, and email fields
// @description - `is_active`: Filter by user active status (true/false)
// @description - `sort_by`: Sort field (id, email, first_name, last_name, is_active, created_at, updated_at)
// @description - `sort_order`: Sort direction (asc, desc)
// @description
// @description **Example:** `GET /users?search=john&is_active=true&sort_by=created_at&sort_order=desc`
// @description
// @description ## 🔐 Security Features
// @description - **JWT Authentication**: Secure token-based authentication
// @description - **Password Hashing**: Bcrypt encryption for user passwords
// @description - **Token Blacklisting**: Secure logout with token invalidation
// @description - **Input Validation**: Comprehensive request validation
// @description - **Rate Limiting**: API rate limiting to prevent abuse
// @description - **CORS Support**: Configurable cross-origin resource sharing
//
// @termsOfService http://swagger.io/terms/
//
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @BasePath /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token. Example: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
//
// @schemes http https
//
// @tag.name Authentication
// @tag.description Authentication endpoints for user registration, login, logout, and token management
//
// @tag.name Users
// @tag.description User management endpoints for profile operations and user listing
package docs
