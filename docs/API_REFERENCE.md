# API Reference Guide

Complete reference for the Go Fiber Template API with detailed examples, request/response formats, and usage patterns.

## 📋 Overview

- **Base URL**: `http://localhost:8080/api/v1`
- **Authentication**: JWT Bearer tokens
- **Content Type**: `application/json`
- **Response Format**: Standardized JSON responses

## 🔐 Authentication

All protected endpoints require a JWT token in the Authorization header:

```
Authorization: Bearer {your-access-token}
```

### Token Types

| Token Type        | Expiry     | Purpose                    |
| ----------------- | ---------- | -------------------------- |
| **Access Token**  | 15 minutes | API requests               |
| **Refresh Token** | 7 days     | Generate new access tokens |

## 📊 Response Format

All API responses follow a consistent structure:

### Success Response

```json
{
  "success": true,
  "message": "Operation successful",
  "data": {
    // Response data here
  },
  "error": null,
  "meta": {
    // Pagination metadata (for list endpoints)
    "page": 1,
    "limit": 10,
    "total": 100,
    "total_pages": 10
  }
}
```

### Error Response

```json
{
  "success": false,
  "message": "Request failed",
  "data": null,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {
      // Field-specific validation errors
      "field_name": "Field-specific error message"
    }
  },
  "meta": null
}
```

## 🔗 Endpoints

### Authentication Endpoints

#### Register User

Create a new user account.

```http
POST /auth/register
```

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Validation Rules:**

- `email`: Valid email format, unique
- `password`: Minimum 8 characters
- `first_name`: 1-100 characters
- `last_name`: 1-100 characters

**Success Response (201):**

```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": 1,
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "is_active": true,
      "created_at": "2025-01-25T10:30:00Z",
      "updated_at": "2025-01-25T10:30:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-01-25T10:45:00Z",
    "token_type": "Bearer"
  }
}
```

**Error Responses:**

- `400` - Validation errors
- `409` - Email already exists
- `500` - Internal server error

#### Login User

Authenticate with existing credentials.

```http
POST /auth/login
```

**Request Body:**

```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Success Response (200):**

```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "user": {
      "id": 1,
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "is_active": true,
      "created_at": "2025-01-25T10:30:00Z",
      "updated_at": "2025-01-25T10:30:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-01-25T10:45:00Z",
    "token_type": "Bearer"
  }
}
```

**Error Responses:**

- `400` - Validation errors
- `401` - Invalid credentials
- `500` - Internal server error

#### Refresh Token

Generate new access token using refresh token.

```http
POST /auth/refresh
```

**Request Body:**

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Success Response (200):**

```json
{
  "success": true,
  "message": "Token refreshed successfully",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-01-25T11:00:00Z",
    "token_type": "Bearer"
  }
}
```

**Error Responses:**

- `400` - Invalid request format
- `401` - Invalid or expired refresh token
- `500` - Internal server error

#### Logout User

Invalidate user tokens.

```http
POST /auth/logout
```

**Headers:**

```
Authorization: Bearer {access-token}
```

**Request Body (Optional):**

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Success Response (200):**

```json
{
  "success": true,
  "message": "Logout successful",
  "data": null
}
```

**Error Responses:**

- `401` - Invalid or missing token
- `500` - Internal server error

### User Management Endpoints

#### Get User Profile

Retrieve user information by ID.

```http
GET /users/{id}
```

**Headers:**

```
Authorization: Bearer {access-token}
```

**Path Parameters:**

- `id` (integer): User ID (minimum: 1)

**Success Response (200):**

```json
{
  "success": true,
  "message": "User profile retrieved successfully",
  "data": {
    "id": 1,
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "is_active": true,
    "created_at": "2025-01-25T10:30:00Z",
    "updated_at": "2025-01-25T10:30:00Z"
  }
}
```

**Error Responses:**

- `401` - Unauthorized
- `404` - User not found
- `500` - Internal server error

#### Update User Profile

Update user information.

```http
PUT /users/{id}
```

**Headers:**

```
Authorization: Bearer {access-token}
```

**Path Parameters:**

- `id` (integer): User ID (minimum: 1)

**Request Body (all fields optional):**

```json
{
  "first_name": "John Updated",
  "last_name": "Doe Updated",
  "is_active": true
}
```

**Validation Rules:**

- `first_name`: 1-100 characters (optional)
- `last_name`: 1-100 characters (optional)
- `is_active`: boolean (optional)

**Success Response (200):**

```json
{
  "success": true,
  "message": "User profile updated successfully",
  "data": {
    "id": 1,
    "email": "user@example.com",
    "first_name": "John Updated",
    "last_name": "Doe Updated",
    "is_active": true,
    "created_at": "2025-01-25T10:30:00Z",
    "updated_at": "2025-01-25T11:00:00Z"
  }
}
```

**Error Responses:**

- `400` - Validation errors
- `401` - Unauthorized
- `404` - User not found
- `500` - Internal server error

#### List Users

Get paginated list of users with filtering and sorting.

```http
GET /users
```

**Headers:**

```
Authorization: Bearer {access-token}
```

**Query Parameters:**

| Parameter    | Type    | Default | Description               |
| ------------ | ------- | ------- | ------------------------- |
| `page`       | integer | 1       | Page number (minimum: 1)  |
| `limit`      | integer | 10      | Items per page (1-100)    |
| `search`     | string  | -       | Search in name/email      |
| `is_active`  | boolean | -       | Filter by active status   |
| `sort_by`    | string  | `id`    | Sort field                |
| `sort_order` | string  | `asc`   | Sort direction (asc/desc) |

**Sort Fields:**

- `id`, `email`, `first_name`, `last_name`, `is_active`, `created_at`, `updated_at`

**Example Requests:**

```http
GET /users
GET /users?page=2&limit=20
GET /users?search=john&is_active=true
GET /users?sort_by=created_at&sort_order=desc
GET /users?page=1&limit=10&search=doe&sort_by=email&sort_order=asc
```

**Success Response (200):**

```json
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": [
    {
      "id": 1,
      "email": "user1@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "is_active": true,
      "created_at": "2025-01-25T10:30:00Z",
      "updated_at": "2025-01-25T10:30:00Z"
    },
    {
      "id": 2,
      "email": "user2@example.com",
      "first_name": "Jane",
      "last_name": "Smith",
      "is_active": true,
      "created_at": "2025-01-25T11:00:00Z",
      "updated_at": "2025-01-25T11:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 25,
    "total_pages": 3
  }
}
```

**Error Responses:**

- `400` - Invalid query parameters
- `401` - Unauthorized
- `500` - Internal server error

#### Delete User

Soft delete a user (marks as deleted, doesn't remove record).

```http
DELETE /users/{id}
```

**Headers:**

```
Authorization: Bearer {access-token}
```

**Path Parameters:**

- `id` (integer): User ID (minimum: 1)

**Success Response (200):**

```json
{
  "success": true,
  "message": "User deleted successfully",
  "data": null
}
```

**Error Responses:**

- `401` - Unauthorized
- `404` - User not found
- `500` - Internal server error

### Health Check Endpoints

#### Application Health

Check application health status.

```http
GET /health
```

**Success Response (200):**

```json
{
  "success": true,
  "message": "Application is healthy",
  "data": {
    "status": "healthy",
    "timestamp": "2025-01-25T10:30:00Z",
    "version": "1.0.0"
  }
}
```

#### Database Health

Check database connectivity.

```http
GET /health/db
```

**Success Response (200):**

```json
{
  "success": true,
  "message": "Database is healthy",
  "data": {
    "status": "healthy",
    "timestamp": "2025-01-25T10:30:00Z",
    "database": "connected"
  }
}
```

**Error Response (503):**

```json
{
  "success": false,
  "message": "Database is unhealthy",
  "data": {
    "status": "unhealthy",
    "timestamp": "2025-01-25T10:30:00Z",
    "database": "disconnected"
  }
}
```

## 🔧 Code Examples

### cURL Examples

#### Complete Authentication Flow

```bash
# 1. Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "SecurePass123!",
    "first_name": "John",
    "last_name": "Doe"
  }'

# 2. Login (if already registered)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "SecurePass123!"
  }'

# 3. Use access token for protected endpoints
ACCESS_TOKEN="your_access_token_here"

# 4. Get user profile
curl -X GET http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 5. Update user profile
curl -X PUT http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John Updated"
  }'

# 6. List users with filtering
curl -X GET "http://localhost:8080/api/v1/users?search=john&is_active=true&sort_by=created_at&sort_order=desc" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# 7. Refresh token
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "your_refresh_token_here"
  }'

# 8. Logout
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "your_refresh_token_here"
  }'
```

### JavaScript/Fetch Examples

#### API Client Class

```javascript
class APIClient {
  constructor(baseURL = "http://localhost:8080/api/v1") {
    this.baseURL = baseURL;
    this.accessToken = localStorage.getItem("access_token");
    this.refreshToken = localStorage.getItem("refresh_token");
  }

  // Set tokens
  setTokens(accessToken, refreshToken) {
    this.accessToken = accessToken;
    this.refreshToken = refreshToken;
    localStorage.setItem("access_token", accessToken);
    localStorage.setItem("refresh_token", refreshToken);
  }

  // Clear tokens
  clearTokens() {
    this.accessToken = null;
    this.refreshToken = null;
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
  }

  // Make authenticated request with automatic token refresh
  async request(endpoint, options = {}) {
    const url = `${this.baseURL}${endpoint}`;

    // Add authorization header if token exists
    if (this.accessToken) {
      options.headers = {
        ...options.headers,
        Authorization: `Bearer ${this.accessToken}`,
      };
    }

    // Add content type for JSON requests
    if (options.body && typeof options.body === "object") {
      options.headers = {
        ...options.headers,
        "Content-Type": "application/json",
      };
      options.body = JSON.stringify(options.body);
    }

    let response = await fetch(url, options);

    // If token expired, try to refresh
    if (response.status === 401 && this.refreshToken) {
      try {
        await this.refreshAccessToken();
        // Retry request with new token
        options.headers["Authorization"] = `Bearer ${this.accessToken}`;
        response = await fetch(url, options);
      } catch (error) {
        this.clearTokens();
        throw new Error("Authentication failed");
      }
    }

    const data = await response.json();

    if (!data.success) {
      throw new Error(data.error?.message || "Request failed");
    }

    return data;
  }

  // Authentication methods
  async register(userData) {
    const data = await this.request("/auth/register", {
      method: "POST",
      body: userData,
    });

    this.setTokens(data.data.access_token, data.data.refresh_token);
    return data.data;
  }

  async login(email, password) {
    const data = await this.request("/auth/login", {
      method: "POST",
      body: { email, password },
    });

    this.setTokens(data.data.access_token, data.data.refresh_token);
    return data.data;
  }

  async refreshAccessToken() {
    const data = await this.request("/auth/refresh", {
      method: "POST",
      body: { refresh_token: this.refreshToken },
    });

    this.accessToken = data.data.access_token;
    localStorage.setItem("access_token", this.accessToken);
    return data.data;
  }

  async logout() {
    try {
      await this.request("/auth/logout", {
        method: "POST",
        body: { refresh_token: this.refreshToken },
      });
    } finally {
      this.clearTokens();
    }
  }

  // User management methods
  async getUser(id) {
    const data = await this.request(`/users/${id}`);
    return data.data;
  }

  async updateUser(id, userData) {
    const data = await this.request(`/users/${id}`, {
      method: "PUT",
      body: userData,
    });
    return data.data;
  }

  async listUsers(params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const endpoint = queryString ? `/users?${queryString}` : "/users";
    const data = await this.request(endpoint);
    return {
      users: data.data,
      meta: data.meta,
    };
  }

  async deleteUser(id) {
    await this.request(`/users/${id}`, {
      method: "DELETE",
    });
  }

  // Health check methods
  async checkHealth() {
    const data = await this.request("/health");
    return data.data;
  }

  async checkDatabaseHealth() {
    const data = await this.request("/health/db");
    return data.data;
  }
}

// Usage example
const api = new APIClient();

// Register and login
try {
  const result = await api.register({
    email: "john.doe@example.com",
    password: "SecurePass123!",
    first_name: "John",
    last_name: "Doe",
  });
  console.log("Registration successful:", result.user);
} catch (error) {
  console.error("Registration failed:", error.message);
}

// List users with pagination and filtering
try {
  const result = await api.listUsers({
    page: 1,
    limit: 10,
    search: "john",
    is_active: true,
    sort_by: "created_at",
    sort_order: "desc",
  });
  console.log("Users:", result.users);
  console.log("Pagination:", result.meta);
} catch (error) {
  console.error("Failed to list users:", error.message);
}
```

### Python Examples

```python
import requests
import json
from typing import Optional, Dict, Any

class APIClient:
    def __init__(self, base_url: str = "http://localhost:8080/api/v1"):
        self.base_url = base_url
        self.access_token: Optional[str] = None
        self.refresh_token: Optional[str] = None

    def set_tokens(self, access_token: str, refresh_token: str):
        self.access_token = access_token
        self.refresh_token = refresh_token

    def clear_tokens(self):
        self.access_token = None
        self.refresh_token = None

    def _get_headers(self) -> Dict[str, str]:
        headers = {"Content-Type": "application/json"}
        if self.access_token:
            headers["Authorization"] = f"Bearer {self.access_token}"
        return headers

    def _request(self, method: str, endpoint: str, data: Optional[Dict] = None) -> Dict[str, Any]:
        url = f"{self.base_url}{endpoint}"
        headers = self._get_headers()

        response = requests.request(
            method=method,
            url=url,
            headers=headers,
            json=data
        )

        # Handle token refresh on 401
        if response.status_code == 401 and self.refresh_token:
            self.refresh_access_token()
            headers = self._get_headers()
            response = requests.request(
                method=method,
                url=url,
                headers=headers,
                json=data
            )

        result = response.json()

        if not result.get("success"):
            error_msg = result.get("error", {}).get("message", "Request failed")
            raise Exception(error_msg)

        return result

    # Authentication methods
    def register(self, email: str, password: str, first_name: str, last_name: str) -> Dict[str, Any]:
        data = {
            "email": email,
            "password": password,
            "first_name": first_name,
            "last_name": last_name
        }
        result = self._request("POST", "/auth/register", data)
        self.set_tokens(result["data"]["access_token"], result["data"]["refresh_token"])
        return result["data"]

    def login(self, email: str, password: str) -> Dict[str, Any]:
        data = {"email": email, "password": password}
        result = self._request("POST", "/auth/login", data)
        self.set_tokens(result["data"]["access_token"], result["data"]["refresh_token"])
        return result["data"]

    def refresh_access_token(self) -> Dict[str, Any]:
        data = {"refresh_token": self.refresh_token}
        result = self._request("POST", "/auth/refresh", data)
        self.access_token = result["data"]["access_token"]
        return result["data"]

    def logout(self):
        try:
            data = {"refresh_token": self.refresh_token}
            self._request("POST", "/auth/logout", data)
        finally:
            self.clear_tokens()

    # User management methods
    def get_user(self, user_id: int) -> Dict[str, Any]:
        result = self._request("GET", f"/users/{user_id}")
        return result["data"]

    def update_user(self, user_id: int, **kwargs) -> Dict[str, Any]:
        result = self._request("PUT", f"/users/{user_id}", kwargs)
        return result["data"]

    def list_users(self, **params) -> Dict[str, Any]:
        query_string = "&".join([f"{k}={v}" for k, v in params.items()])
        endpoint = f"/users?{query_string}" if query_string else "/users"
        result = self._request("GET", endpoint)
        return {"users": result["data"], "meta": result["meta"]}

    def delete_user(self, user_id: int):
        self._request("DELETE", f"/users/{user_id}")

# Usage example
api = APIClient()

# Register and login
try:
    result = api.register(
        email="john.doe@example.com",
        password="SecurePass123!",
        first_name="John",
        last_name="Doe"
    )
    print("Registration successful:", result["user"])
except Exception as e:
    print("Registration failed:", str(e))

# List users with filtering
try:
    result = api.list_users(
        page=1,
        limit=10,
        search="john",
        is_active=True,
        sort_by="created_at",
        sort_order="desc"
    )
    print("Users:", result["users"])
    print("Pagination:", result["meta"])
except Exception as e:
    print("Failed to list users:", str(e))
```

## 📊 HTTP Status Codes

| Code  | Status                | Description                            |
| ----- | --------------------- | -------------------------------------- |
| `200` | OK                    | Request successful                     |
| `201` | Created               | Resource created successfully          |
| `400` | Bad Request           | Invalid request or validation errors   |
| `401` | Unauthorized          | Invalid/missing/expired authentication |
| `403` | Forbidden             | Access denied                          |
| `404` | Not Found             | Resource not found                     |
| `409` | Conflict              | Resource already exists                |
| `500` | Internal Server Error | Server error                           |
| `503` | Service Unavailable   | Service temporarily unavailable        |

## 🚨 Error Codes

| Code               | Description               | Common Causes                                 |
| ------------------ | ------------------------- | --------------------------------------------- |
| `VALIDATION_ERROR` | Request validation failed | Invalid email format, missing required fields |
| `UNAUTHORIZED`     | Authentication failed     | Invalid credentials, expired token            |
| `NOT_FOUND`        | Resource not found        | Invalid user ID, deleted resource             |
| `CONFLICT`         | Resource conflict         | Email already registered                      |
| `INTERNAL_ERROR`   | Server error              | Database connection, unexpected errors        |

## 🔒 Security Considerations

1. **HTTPS Only**: Always use HTTPS in production
2. **Token Storage**: Store tokens securely (httpOnly cookies recommended)
3. **Token Expiration**: Implement automatic token refresh
4. **Rate Limiting**: API includes built-in rate limiting
5. **Input Validation**: All inputs are validated server-side
6. **SQL Injection**: Protected by GORM parameterized queries
7. **CORS**: Configurable CORS settings

## 📚 Additional Resources

- **[Authentication Examples](examples/authentication.md)** - Detailed auth flow examples
- **[User Management Examples](examples/users.md)** - User CRUD operations
- **[API Testing Guide](examples/api-testing-guide.md)** - Complete testing workflows
- **[Swagger UI](http://localhost:8080/swagger/)** - Interactive API documentation

---

**💡 Pro Tip**: Use the interactive Swagger UI at http://localhost:8080/swagger/ for the best API exploration and testing experience!
