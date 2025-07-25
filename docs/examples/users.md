# User Management Examples

This document provides detailed examples for all user management endpoints in the Go Fiber Template API.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication Required

All user management endpoints require a valid JWT token in the Authorization header:

```
Authorization: Bearer {your-jwt-token}
```

## User Management Operations

### 1. Get User Profile

Retrieve user profile information by ID.

**Endpoint:** `GET /users/{id}`

**Request Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example_token
Content-Type: application/json
```

**Path Parameters:**

- `id` (integer, required): User ID (minimum: 1)

**Success Response (200 OK):**

```json
{
  "success": true,
  "message": "User profile retrieved successfully",
  "data": {
    "id": 1,
    "email": "john.doe@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "is_active": true,
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  }
}
```

**Error Response (404 Not Found):**

```json
{
  "success": false,
  "message": "Request failed",
  "error": {
    "code": "NOT_FOUND",
    "message": "User not found"
  }
}
```

### 2. Update User Profile

Update user profile information. All fields are optional.

**Endpoint:** `PUT /users/{id}`

**Request Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example_token
Content-Type: application/json
```

**Path Parameters:**

- `id` (integer, required): User ID (minimum: 1)

**Request Body:**

```json
{
  "first_name": "John Updated",
  "last_name": "Doe Updated",
  "is_active": true
}
```

**Request Body (Partial Update):**

```json
{
  "first_name": "John Updated"
}
```

**Success Response (200 OK):**

```json
{
  "success": true,
  "message": "User profile updated successfully",
  "data": {
    "id": 1,
    "email": "john.doe@example.com",
    "first_name": "John Updated",
    "last_name": "Doe Updated",
    "is_active": true,
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T12:00:00Z"
  }
}
```

**Error Response (400 Bad Request - Validation Error):**

```json
{
  "success": false,
  "message": "Request failed",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": {
      "first_name": "First name must be between 1 and 100 characters"
    }
  }
}
```

### 3. List Users

Get a paginated list of users with optional filtering and sorting.

**Endpoint:** `GET /users`

**Request Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example_token
Content-Type: application/json
```

**Query Parameters:**

- `page` (integer, optional): Page number (default: 1, minimum: 1)
- `limit` (integer, optional): Items per page (default: 10, minimum: 1, maximum: 100)
- `search` (string, optional): Search term for name or email
- `is_active` (boolean, optional): Filter by active status
- `sort_by` (string, optional): Sort field (id, email, first_name, last_name, is_active, created_at, updated_at)
- `sort_order` (string, optional): Sort order (asc, desc)

**Example Request URLs:**

```
GET /users
GET /users?page=2&limit=20
GET /users?search=john&is_active=true
GET /users?sort_by=created_at&sort_order=desc
GET /users?page=1&limit=10&search=doe&sort_by=email&sort_order=asc
```

**Success Response (200 OK):**

```json
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": [
    {
      "id": 1,
      "email": "john.doe@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "is_active": true,
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z"
    },
    {
      "id": 2,
      "email": "jane.smith@example.com",
      "first_name": "Jane",
      "last_name": "Smith",
      "is_active": true,
      "created_at": "2023-01-02T00:00:00Z",
      "updated_at": "2023-01-02T00:00:00Z"
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

**Success Response (Empty Results):**

```json
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": [],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 0,
    "total_pages": 0
  }
}
```

### 4. Delete User

Soft delete a user by ID. This performs a soft delete, marking the user as deleted without removing the record.

**Endpoint:** `DELETE /users/{id}`

**Request Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example_token
Content-Type: application/json
```

**Path Parameters:**

- `id` (integer, required): User ID (minimum: 1)

**Success Response (200 OK):**

```json
{
  "success": true,
  "message": "User deleted successfully"
}
```

**Error Response (404 Not Found):**

```json
{
  "success": false,
  "message": "Request failed",
  "error": {
    "code": "NOT_FOUND",
    "message": "User not found"
  }
}
```

## cURL Examples

### Get user profile

```bash
curl -X GET http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer your_access_token_here" \
  -H "Content-Type: application/json"
```

### Update user profile

```bash
curl -X PUT http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer your_access_token_here" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John Updated",
    "last_name": "Doe Updated"
  }'
```

### List users with pagination

```bash
curl -X GET "http://localhost:8080/api/v1/users?page=1&limit=10" \
  -H "Authorization: Bearer your_access_token_here" \
  -H "Content-Type: application/json"
```

### List users with search and filtering

```bash
curl -X GET "http://localhost:8080/api/v1/users?search=john&is_active=true&sort_by=created_at&sort_order=desc" \
  -H "Authorization: Bearer your_access_token_here" \
  -H "Content-Type: application/json"
```

### Delete user

```bash
curl -X DELETE http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer your_access_token_here" \
  -H "Content-Type: application/json"
```

## JavaScript/Fetch Examples

### Get user profile

```javascript
const accessToken = localStorage.getItem("access_token");

const response = await fetch("http://localhost:8080/api/v1/users/1", {
  method: "GET",
  headers: {
    Authorization: `Bearer ${accessToken}`,
    "Content-Type": "application/json",
  },
});

const data = await response.json();
if (data.success) {
  console.log("User profile:", data.data);
} else {
  console.error("Error:", data.error);
}
```

### Update user profile

```javascript
const accessToken = localStorage.getItem("access_token");

const response = await fetch("http://localhost:8080/api/v1/users/1", {
  method: "PUT",
  headers: {
    Authorization: `Bearer ${accessToken}`,
    "Content-Type": "application/json",
  },
  body: JSON.stringify({
    first_name: "John Updated",
    last_name: "Doe Updated",
  }),
});

const data = await response.json();
if (data.success) {
  console.log("Updated user:", data.data);
} else {
  console.error("Error:", data.error);
}
```

### List users with pagination

```javascript
const accessToken = localStorage.getItem("access_token");

const params = new URLSearchParams({
  page: "1",
  limit: "10",
  search: "john",
  is_active: "true",
  sort_by: "created_at",
  sort_order: "desc",
});

const response = await fetch(`http://localhost:8080/api/v1/users?${params}`, {
  method: "GET",
  headers: {
    Authorization: `Bearer ${accessToken}`,
    "Content-Type": "application/json",
  },
});

const data = await response.json();
if (data.success) {
  console.log("Users:", data.data);
  console.log("Pagination:", data.meta);
} else {
  console.error("Error:", data.error);
}
```

### Delete user

```javascript
const accessToken = localStorage.getItem("access_token");

const response = await fetch("http://localhost:8080/api/v1/users/1", {
  method: "DELETE",
  headers: {
    Authorization: `Bearer ${accessToken}`,
    "Content-Type": "application/json",
  },
});

const data = await response.json();
if (data.success) {
  console.log("User deleted successfully");
} else {
  console.error("Error:", data.error);
}
```

## Error Handling

### Common Error Responses

**401 Unauthorized (Invalid/Missing Token):**

```json
{
  "success": false,
  "message": "Request failed",
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or missing JWT token"
  }
}
```

**400 Bad Request (Invalid Parameters):**

```json
{
  "success": false,
  "message": "Request failed",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid query parameters"
  }
}
```

**500 Internal Server Error:**

```json
{
  "success": false,
  "message": "Request failed",
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Internal server error"
  }
}
```

## Pagination Guidelines

### Query Parameters

- `page`: Page number starting from 1
- `limit`: Number of items per page (1-100)

### Response Metadata

The `meta` object in paginated responses contains:

- `page`: Current page number
- `limit`: Items per page
- `total`: Total number of items
- `total_pages`: Total number of pages

### Navigation Examples

```javascript
// First page
GET /users?page=1&limit=10

// Next page
GET /users?page=2&limit=10

// Last page (if total_pages = 5)
GET /users?page=5&limit=10
```

## Search and Filtering

### Search

The `search` parameter performs a case-insensitive search across:

- First name
- Last name
- Email address

### Filtering

- `is_active`: Filter by user active status (true/false)

### Sorting

- `sort_by`: Field to sort by (id, email, first_name, last_name, is_active, created_at, updated_at)
- `sort_order`: Sort direction (asc, desc)

### Combined Example

```
GET /users?search=john&is_active=true&sort_by=created_at&sort_order=desc&page=1&limit=20
```

This searches for users with "john" in their name/email, filters for active users only, sorts by creation date (newest first), and returns the first page with 20 items.
