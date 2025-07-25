# API Response Format

This document describes the standardized response format used throughout the Go Fiber Template API.

## Overview

All API endpoints return responses in a consistent JSON structure that includes status information, data, error details, and metadata. The response format uses integer codes for easy programmatic handling while maintaining string codes in error details for backward compatibility.

## Response Structure

### Base Response Format

```json
{
  "success": boolean,
  "code": integer,
  "message": string,
  "data": object|array|null,
  "error": object|null,
  "meta": object|null
}
```

### Field Descriptions

- **`success`** (boolean): Indicates whether the request was successful
- **`code`** (integer): HTTP status code for easy programmatic handling
- **`message`** (string): Human-readable message describing the result
- **`data`** (object|array|null): Response payload (present in successful responses)
- **`error`** (object|null): Error details (present only in error responses)
- **`meta`** (object|null): Additional metadata like pagination info

## Response Codes

### Success Codes

- **200** - Success: Operation completed successfully
- **201** - Created: Resource created successfully

### Client Error Codes

- **400** - Bad Request: Invalid request format or parameters
- **401** - Unauthorized: Invalid credentials or expired token
- **403** - Forbidden: Access denied
- **404** - Not Found: Resource not found
- **422** - Validation Error: Request validation failed
- **429** - Too Many Requests: Rate limit exceeded

### Server Error Codes

- **500** - Internal Server Error: Server error

## Response Examples

### Successful Registration

```json
{
  "success": true,
  "code": 201,
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": 1,
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "is_active": true,
      "created_at": "2025-07-25T06:13:59Z",
      "updated_at": "2025-07-25T06:13:59Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  },
  "error": null,
  "meta": null
}
```

### Successful Login

```json
{
  "success": true,
  "code": 200,
  "message": "Login successful",
  "data": {
    "user": {
      "id": 1,
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "is_active": true,
      "created_at": "2025-07-25T06:13:59Z",
      "updated_at": "2025-07-25T06:13:59Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  },
  "error": null,
  "meta": null
}
```

### Paginated Response (User List)

```json
{
  "success": true,
  "code": 200,
  "message": "Users retrieved successfully",
  "data": [
    {
      "id": 1,
      "email": "user1@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "is_active": true,
      "created_at": "2025-07-25T06:13:59Z",
      "updated_at": "2025-07-25T06:13:59Z"
    },
    {
      "id": 2,
      "email": "user2@example.com",
      "first_name": "Jane",
      "last_name": "Smith",
      "is_active": true,
      "created_at": "2025-07-25T06:14:30Z",
      "updated_at": "2025-07-25T06:14:30Z"
    }
  ],
  "error": null,
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 25,
    "total_pages": 3
  }
}
```

### Validation Error Response

```json
{
  "success": false,
  "code": 422,
  "message": "Request failed",
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": {
      "email": "Email must be a valid email address",
      "password": "Password must be at least 8 characters long",
      "first_name": "First name is required"
    }
  },
  "meta": null
}
```

### Unauthorized Error Response

```json
{
  "success": false,
  "code": 401,
  "message": "Request failed",
  "data": null,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid credentials",
    "details": null
  },
  "meta": null
}
```

### Not Found Error Response

```json
{
  "success": false,
  "code": 404,
  "message": "Request failed",
  "data": null,
  "error": {
    "code": "NOT_FOUND",
    "message": "User not found",
    "details": null
  },
  "meta": null
}
```

### Rate Limit Error Response

```json
{
  "success": false,
  "code": 429,
  "message": "Request failed",
  "data": null,
  "error": {
    "code": "TOO_MANY_REQUESTS",
    "message": "Rate limit exceeded. Please try again later.",
    "details": null
  },
  "meta": null
}
```

### Internal Server Error Response

```json
{
  "success": false,
  "code": 500,
  "message": "Request failed",
  "data": null,
  "error": {
    "code": "INTERNAL_SERVER_ERROR",
    "message": "An unexpected error occurred",
    "details": null
  },
  "meta": null
}
```

## Error Structure

When an error occurs, the `error` field contains:

```json
{
  "code": "ERROR_CODE_STRING",
  "message": "Human-readable error message",
  "details": {
    "field1": "Field-specific error message",
    "field2": "Another field-specific error message"
  }
}
```

### Error Code Mapping

| Integer Code | String Code           | Description                          |
| ------------ | --------------------- | ------------------------------------ |
| 400          | BAD_REQUEST           | Invalid request format or parameters |
| 401          | UNAUTHORIZED          | Invalid credentials or expired token |
| 403          | FORBIDDEN             | Access denied                        |
| 404          | NOT_FOUND             | Resource not found                   |
| 422          | VALIDATION_ERROR      | Request validation failed            |
| 429          | TOO_MANY_REQUESTS     | Rate limit exceeded                  |
| 500          | INTERNAL_SERVER_ERROR | Server error                         |

## Pagination Metadata

For paginated responses, the `meta` field contains:

```json
{
  "page": 1,
  "limit": 10,
  "total": 100,
  "total_pages": 10
}
```

- **`page`**: Current page number (1-based)
- **`limit`**: Number of items per page
- **`total`**: Total number of items available
- **`total_pages`**: Total number of pages available

## Frontend Integration

### JavaScript Example

```javascript
// Handle API response
function handleApiResponse(response) {
  if (response.success) {
    switch (response.code) {
      case 200:
        console.log("Success:", response.message);
        return response.data;
      case 201:
        console.log("Created:", response.message);
        return response.data;
      default:
        console.log("Unexpected success code:", response.code);
        return response.data;
    }
  } else {
    switch (response.code) {
      case 400:
        console.error("Bad Request:", response.error.message);
        break;
      case 401:
        console.error("Unauthorized:", response.error.message);
        // Redirect to login
        break;
      case 403:
        console.error("Forbidden:", response.error.message);
        break;
      case 404:
        console.error("Not Found:", response.error.message);
        break;
      case 422:
        console.error("Validation Error:", response.error.details);
        // Handle field-specific errors
        break;
      case 429:
        console.error("Rate Limited:", response.error.message);
        break;
      case 500:
        console.error("Server Error:", response.error.message);
        break;
      default:
        console.error("Unknown error:", response.error.message);
    }
    throw new Error(response.error.message);
  }
}

// Usage example
fetch("/api/v1/auth/login", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ email: "user@example.com", password: "password" }),
})
  .then((response) => response.json())
  .then(handleApiResponse)
  .then((data) => {
    // Handle successful response data
    console.log("User:", data.user);
    localStorage.setItem("access_token", data.access_token);
  })
  .catch((error) => {
    // Handle error
    console.error("Login failed:", error.message);
  });
```

## Best Practices

1. **Always check the `success` field** first to determine if the request was successful
2. **Use the integer `code` field** for programmatic handling and routing
3. **Use the string `error.code` field** for detailed error handling and user messaging
4. **Handle validation errors** by iterating through `error.details` for field-specific messages
5. **Implement proper error handling** for all possible response codes
6. **Use pagination metadata** to implement proper pagination controls
7. **Store and use JWT tokens** from successful authentication responses

## Testing

The response format is thoroughly tested in `internal/utils/response_test.go`. Run tests with:

```bash
go test ./internal/utils -v
```

All response functions are tested to ensure they return the correct format with appropriate codes and messages.
