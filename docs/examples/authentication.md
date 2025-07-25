# Authentication Examples

This document provides detailed examples for using the authentication endpoints in the Go Fiber Template API.

## Base URL

All API endpoints are available at: `http://localhost:8080/api/v1`

## Authentication Flow

### 1. User Registration

Register a new user account with email and password.

**Endpoint:** `POST /auth/register`

**Request Body:**

```json
{
  "email": "john.doe@example.com",
  "password": "securepassword123",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Success Response (201):**

```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": 1,
      "email": "john.doe@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "is_active": true,
      "created_at": "2025-07-24T10:30:00Z",
      "updated_at": "2025-07-24T10:30:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1721819400,
    "token_type": "Bearer"
  },
  "error": null,
  "meta": null
}
```

**Error Response (400) - Validation Error:**

```json
{
  "success": false,
  "message": "Request failed",
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "details": {
      "email": "Email is required and must be valid",
      "password": "Password must be at least 8 characters long"
    }
  },
  "meta": null
}
```

**Error Response (409) - Email Already Exists:**

```json
{
  "success": false,
  "message": "Request failed",
  "data": null,
  "error": {
    "code": "CONFLICT",
    "message": "Email already exists",
    "details": null
  },
  "meta": null
}
```

### 2. User Login

Authenticate with existing credentials to get access tokens.

**Endpoint:** `POST /auth/login`

**Request Body:**

```json
{
  "email": "john.doe@example.com",
  "password": "securepassword123"
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
      "email": "john.doe@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "is_active": true,
      "created_at": "2025-07-24T10:30:00Z",
      "updated_at": "2025-07-24T10:30:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": 1721819400,
    "token_type": "Bearer"
  },
  "error": null,
  "meta": null
}
```

**Error Response (401) - Invalid Credentials:**

```json
{
  "success": false,
  "message": "Request failed",
  "data": null,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid email or password",
    "details": null
  },
  "meta": null
}
```

### 3. Token Refresh

Generate a new access token using a valid refresh token.

**Endpoint:** `POST /auth/refresh`

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
    "expires_at": 1721820300,
    "token_type": "Bearer"
  },
  "error": null,
  "meta": null
}
```

**Error Response (401) - Invalid Refresh Token:**

```json
{
  "success": false,
  "message": "Request failed",
  "data": null,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or expired refresh token",
    "details": null
  },
  "meta": null
}
```

### 4. User Logout

Invalidate tokens and logout the user.

**Endpoint:** `POST /auth/logout`

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
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
  "data": null,
  "error": null,
  "meta": null
}
```

**Error Response (401) - Invalid Token:**

```json
{
  "success": false,
  "message": "Request failed",
  "data": null,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid authorization header",
    "details": null
  },
  "meta": null
}
```

## cURL Examples

### Register a new user

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "securepassword123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### Login with credentials

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "securepassword123"
  }'
```

### Refresh access token

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "your_refresh_token_here"
  }'
```

### Logout user

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer your_access_token_here" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "your_refresh_token_here"
  }'
```

## JavaScript/Fetch Examples

### Register a new user

```javascript
const registerUser = async () => {
  const response = await fetch("http://localhost:8080/api/v1/auth/register", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      email: "john.doe@example.com",
      password: "securepassword123",
      first_name: "John",
      last_name: "Doe",
    }),
  });

  const data = await response.json();

  if (data.success) {
    // Store tokens for later use
    localStorage.setItem("access_token", data.data.access_token);
    localStorage.setItem("refresh_token", data.data.refresh_token);
    console.log("Registration successful:", data.data.user);
  } else {
    console.error("Registration failed:", data.error);
  }
};
```

### Login user

```javascript
const loginUser = async (email, password) => {
  const response = await fetch("http://localhost:8080/api/v1/auth/login", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ email, password }),
  });

  const data = await response.json();

  if (data.success) {
    // Store tokens for later use
    localStorage.setItem("access_token", data.data.access_token);
    localStorage.setItem("refresh_token", data.data.refresh_token);
    console.log("Login successful:", data.data.user);
    return data.data;
  } else {
    console.error("Login failed:", data.error);
    throw new Error(data.error.message);
  }
};
```

### Refresh token

```javascript
const refreshToken = async () => {
  const refreshToken = localStorage.getItem("refresh_token");

  const response = await fetch("http://localhost:8080/api/v1/auth/refresh", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });

  const data = await response.json();

  if (data.success) {
    // Update stored access token
    localStorage.setItem("access_token", data.data.access_token);
    console.log("Token refreshed successfully");
    return data.data.access_token;
  } else {
    console.error("Token refresh failed:", data.error);
    // Clear tokens and redirect to login
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
    throw new Error(data.error.message);
  }
};
```

### Logout user

```javascript
const logoutUser = async () => {
  const accessToken = localStorage.getItem("access_token");
  const refreshToken = localStorage.getItem("refresh_token");

  const response = await fetch("http://localhost:8080/api/v1/auth/logout", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${accessToken}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });

  const data = await response.json();

  // Clear tokens regardless of response
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");

  if (data.success) {
    console.log("Logout successful");
  } else {
    console.error("Logout failed:", data.error);
  }
};
```

## Authentication Helper Functions

### Automatic token refresh

```javascript
const makeAuthenticatedRequest = async (url, options = {}) => {
  let accessToken = localStorage.getItem("access_token");

  // Add authorization header
  options.headers = {
    ...options.headers,
    Authorization: `Bearer ${accessToken}`,
    "Content-Type": "application/json",
  };

  let response = await fetch(url, options);

  // If token expired, try to refresh
  if (response.status === 401) {
    try {
      accessToken = await refreshToken();
      options.headers["Authorization"] = `Bearer ${accessToken}`;
      response = await fetch(url, options);
    } catch (error) {
      // Refresh failed, redirect to login
      window.location.href = "/login";
      return;
    }
  }

  return response;
};
```

### Check if user is authenticated

```javascript
const isAuthenticated = () => {
  const accessToken = localStorage.getItem("access_token");
  const refreshToken = localStorage.getItem("refresh_token");

  return !!(accessToken && refreshToken);
};
```

### Parse JWT token (client-side only for UI purposes)

```javascript
const parseJWT = (token) => {
  try {
    const base64Url = token.split(".")[1];
    const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split("")
        .map(function (c) {
          return "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2);
        })
        .join("")
    );

    return JSON.parse(jsonPayload);
  } catch (error) {
    return null;
  }
};

const isTokenExpired = (token) => {
  const payload = parseJWT(token);
  if (!payload) return true;

  const currentTime = Date.now() / 1000;
  return payload.exp < currentTime;
};
```

## Security Best Practices

1. **Store tokens securely**: Use httpOnly cookies in production instead of localStorage
2. **Validate tokens server-side**: Never trust client-side token validation
3. **Use HTTPS**: Always use HTTPS in production to protect tokens in transit
4. **Implement token rotation**: Regularly refresh access tokens
5. **Handle token expiration**: Implement automatic token refresh with fallback to login
6. **Logout on tab close**: Consider implementing logout on browser/tab close
7. **Rate limiting**: The API includes rate limiting to prevent brute force attacks

## Common Error Codes

- `VALIDATION_ERROR`: Request validation failed
- `UNAUTHORIZED`: Invalid credentials or expired token
- `CONFLICT`: Resource already exists (e.g., email already registered)
- `INTERNAL_ERROR`: Server error occurred

## Token Expiration Times

- **Access Token**: 15 minutes
- **Refresh Token**: 7 days (168 hours)

Access tokens should be refreshed before expiration using the refresh token. When the refresh token expires, the user must log in again.
