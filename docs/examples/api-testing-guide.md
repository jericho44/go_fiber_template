# API Testing Guide

This comprehensive guide provides detailed instructions for testing the Go Fiber Template API endpoints using various tools and methods, with a focus on the interactive Swagger UI.

## Prerequisites

1. **Server Running**: Ensure the API server is running on `http://localhost:8080`
2. **Database**: PostgreSQL database should be set up and migrations applied
3. **Environment**: Proper environment variables configured (see `.env.example`)

## Testing Tools

### 1. Swagger UI (Recommended) 🚀

The **easiest and most comprehensive** way to test the API is through the built-in Swagger UI with complete interactive documentation.

**Access Swagger UI:**

```
http://localhost:8080/swagger/
```

**Key Features:**

- 📖 **Interactive API Documentation** - Complete endpoint documentation with examples
- 🧪 **Built-in Testing** - Test endpoints directly from the browser
- 🔐 **Authentication Support** - Easy JWT token management
- 📋 **Request/Response Examples** - See real request and response formats
- 🔍 **Schema Validation** - Automatic request validation
- 📊 **Real-time Testing** - Immediate feedback on API calls

#### Complete Swagger UI Testing Workflow

**Step 1: Open Swagger UI**

1. Navigate to `http://localhost:8080/swagger/` in your browser
2. You'll see the complete API documentation with two main sections:
   - **Authentication** - User registration, login, logout, token refresh
   - **Users** - User management operations

**Step 2: Test Authentication Flow**

1. **Register a New User:**

   - Expand the **Authentication** section
   - Click on `POST /auth/register`
   - Click **"Try it out"**
   - Fill in the request body:
     ```json
     {
       "email": "test@example.com",
       "password": "testpassword123",
       "first_name": "Test",
       "last_name": "User"
     }
     ```
   - Click **"Execute"**
   - **Copy the `access_token`** from the response data

2. **Authorize for Protected Endpoints:**
   - Click the **"Authorize"** button at the top of the page (🔒 icon)
   - In the popup, enter: `Bearer {your-access-token}` (replace with your actual token)
   - Click **"Authorize"**
   - Click **"Close"**
   - You'll now see a 🔒 icon next to protected endpoints

**Step 3: Test Protected User Endpoints**

1. **List Users:**

   - Expand the **Users** section
   - Click on `GET /users`
   - Click **"Try it out"**
   - Optionally modify query parameters (page, limit, search, etc.)
   - Click **"Execute"**
   - View the paginated user list response

2. **Get User Profile:**

   - Click on `GET /users/{id}`
   - Click **"Try it out"**
   - Enter a user ID (e.g., `1`)
   - Click **"Execute"**
   - View the user profile response

3. **Update User Profile:**
   - Click on `PUT /users/{id}`
   - Click **"Try it out"**
   - Enter a user ID
   - Modify the request body:
     ```json
     {
       "first_name": "Updated Name",
       "last_name": "Updated Last"
     }
     ```
   - Click **"Execute"**
   - View the updated user response

**Step 4: Test Token Management**

1. **Refresh Token:**

   - Go back to **Authentication** section
   - Click on `POST /auth/refresh`
   - Click **"Try it out"**
   - Enter your refresh token from the login response:
     ```json
     {
       "refresh_token": "your_refresh_token_here"
     }
     ```
   - Click **"Execute"**
   - Get a new access token

2. **Logout:**
   - Click on `POST /auth/logout`
   - Click **"Try it out"**
   - Optionally add refresh token to request body
   - Click **"Execute"**
   - Your tokens are now invalidated

**Step 5: Test Error Scenarios**

1. **Test Invalid Login:**

   - Try `POST /auth/login` with wrong credentials
   - Observe the 401 error response format

2. **Test Unauthorized Access:**

   - Click "Authorize" and clear your token
   - Try accessing `GET /users`
   - Observe the 401 unauthorized response

3. **Test Validation Errors:**
   - Try `POST /auth/register` with invalid email format
   - Observe the 400 validation error with field details

#### Swagger UI Pro Tips

- **Response Examples**: Each endpoint shows example responses for different status codes
- **Schema Documentation**: Click on schema names to see detailed field descriptions
- **Try Different Parameters**: Test pagination, search, and filtering options
- **Copy cURL Commands**: Swagger generates cURL commands you can copy and use
- **Download OpenAPI Spec**: Use the `/swagger/doc.json` endpoint to get the raw OpenAPI specification

### 2. cURL Commands

#### Authentication Flow Test

```bash
# 1. Register a new user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "testpassword123",
    "first_name": "Test",
    "last_name": "User"
  }'

# 2. Login (save the access_token from response)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "testpassword123"
  }'

# 3. Use the access token for protected endpoints
ACCESS_TOKEN="your_access_token_here"

# 4. Get user profile
curl -X GET http://localhost:8080/api/v1/users/1 \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json"

# 5. List users
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json"

# 6. Logout
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json"
```

### 3. Postman Collection

Create a Postman collection for comprehensive API testing:

#### Collection Structure:

```
Go Fiber Template API/
├── Authentication/
│   ├── Register User
│   ├── Login User
│   ├── Refresh Token
│   └── Logout User
└── User Management/
    ├── Get User Profile
    ├── Update User Profile
    ├── List Users
    └── Delete User
```

#### Environment Variables:

```json
{
  "base_url": "http://localhost:8080/api/v1",
  "access_token": "",
  "refresh_token": "",
  "user_id": ""
}
```

#### Pre-request Scripts (for authentication):

```javascript
// For protected endpoints, add this pre-request script:
if (pm.environment.get("access_token")) {
  pm.request.headers.add({
    key: "Authorization",
    value: "Bearer " + pm.environment.get("access_token"),
  });
}
```

#### Test Scripts (for login endpoint):

```javascript
// Save tokens after successful login
if (pm.response.code === 200) {
  const response = pm.response.json();
  if (response.success && response.data) {
    pm.environment.set("access_token", response.data.access_token);
    pm.environment.set("refresh_token", response.data.refresh_token);
    pm.environment.set("user_id", response.data.user.id);
  }
}
```

### 4. JavaScript/Node.js Testing

#### Complete Test Suite Example:

```javascript
const BASE_URL = "http://localhost:8080/api/v1";

class APITester {
  constructor() {
    this.accessToken = null;
    this.refreshToken = null;
    this.userId = null;
  }

  async request(method, endpoint, data = null, requiresAuth = false) {
    const headers = {
      "Content-Type": "application/json",
    };

    if (requiresAuth && this.accessToken) {
      headers["Authorization"] = `Bearer ${this.accessToken}`;
    }

    const config = {
      method,
      headers,
    };

    if (data) {
      config.body = JSON.stringify(data);
    }

    const response = await fetch(`${BASE_URL}${endpoint}`, config);
    const result = await response.json();

    console.log(`${method} ${endpoint}:`, response.status, result);
    return { status: response.status, data: result };
  }

  async testAuthenticationFlow() {
    console.log("=== Testing Authentication Flow ===");

    // 1. Register
    const registerData = {
      email: `test${Date.now()}@example.com`,
      password: "testpassword123",
      first_name: "Test",
      last_name: "User",
    };

    const registerResult = await this.request(
      "POST",
      "/auth/register",
      registerData
    );
    if (registerResult.status === 201) {
      this.accessToken = registerResult.data.data.access_token;
      this.refreshToken = registerResult.data.data.refresh_token;
      this.userId = registerResult.data.data.user.id;
      console.log("✅ Registration successful");
    } else {
      console.log("❌ Registration failed");
      return;
    }

    // 2. Login
    const loginData = {
      email: registerData.email,
      password: registerData.password,
    };

    const loginResult = await this.request("POST", "/auth/login", loginData);
    if (loginResult.status === 200) {
      this.accessToken = loginResult.data.data.access_token;
      console.log("✅ Login successful");
    } else {
      console.log("❌ Login failed");
    }

    // 3. Refresh Token
    const refreshResult = await this.request("POST", "/auth/refresh", {
      refresh_token: this.refreshToken,
    });
    if (refreshResult.status === 200) {
      this.accessToken = refreshResult.data.data.access_token;
      console.log("✅ Token refresh successful");
    } else {
      console.log("❌ Token refresh failed");
    }
  }

  async testUserManagement() {
    console.log("=== Testing User Management ===");

    // 1. Get User Profile
    const profileResult = await this.request(
      "GET",
      `/users/${this.userId}`,
      null,
      true
    );
    if (profileResult.status === 200) {
      console.log("✅ Get user profile successful");
    } else {
      console.log("❌ Get user profile failed");
    }

    // 2. Update User Profile
    const updateData = {
      first_name: "Updated Test",
      last_name: "Updated User",
    };

    const updateResult = await this.request(
      "PUT",
      `/users/${this.userId}`,
      updateData,
      true
    );
    if (updateResult.status === 200) {
      console.log("✅ Update user profile successful");
    } else {
      console.log("❌ Update user profile failed");
    }

    // 3. List Users
    const listResult = await this.request(
      "GET",
      "/users?page=1&limit=10",
      null,
      true
    );
    if (listResult.status === 200) {
      console.log("✅ List users successful");
    } else {
      console.log("❌ List users failed");
    }
  }

  async testErrorHandling() {
    console.log("=== Testing Error Handling ===");

    // 1. Invalid login
    const invalidLogin = await this.request("POST", "/auth/login", {
      email: "invalid@example.com",
      password: "wrongpassword",
    });
    if (invalidLogin.status === 401) {
      console.log("✅ Invalid login properly rejected");
    } else {
      console.log("❌ Invalid login not properly handled");
    }

    // 2. Unauthorized access
    const unauthorizedAccess = await this.request(
      "GET",
      "/users/1",
      null,
      false
    );
    if (unauthorizedAccess.status === 401) {
      console.log("✅ Unauthorized access properly rejected");
    } else {
      console.log("❌ Unauthorized access not properly handled");
    }

    // 3. Invalid user ID
    const invalidUser = await this.request("GET", "/users/99999", null, true);
    if (invalidUser.status === 404) {
      console.log("✅ Invalid user ID properly handled");
    } else {
      console.log("❌ Invalid user ID not properly handled");
    }
  }

  async logout() {
    console.log("=== Logging Out ===");
    const logoutResult = await this.request(
      "POST",
      "/auth/logout",
      {
        refresh_token: this.refreshToken,
      },
      true
    );

    if (logoutResult.status === 200) {
      console.log("✅ Logout successful");
    } else {
      console.log("❌ Logout failed");
    }
  }

  async runAllTests() {
    try {
      await this.testAuthenticationFlow();
      await this.testUserManagement();
      await this.testErrorHandling();
      await this.logout();
      console.log("=== All Tests Completed ===");
    } catch (error) {
      console.error("Test suite failed:", error);
    }
  }
}

// Run the tests
const tester = new APITester();
tester.runAllTests();
```

### 5. Python Testing with requests

```python
import requests
import json
import time

class APITester:
    def __init__(self, base_url="http://localhost:8080/api/v1"):
        self.base_url = base_url
        self.access_token = None
        self.refresh_token = None
        self.user_id = None

    def request(self, method, endpoint, data=None, requires_auth=False):
        headers = {"Content-Type": "application/json"}

        if requires_auth and self.access_token:
            headers["Authorization"] = f"Bearer {self.access_token}"

        url = f"{self.base_url}{endpoint}"

        response = requests.request(
            method=method,
            url=url,
            headers=headers,
            json=data if data else None
        )

        print(f"{method} {endpoint}: {response.status_code}")

        try:
            result = response.json()
            print(json.dumps(result, indent=2))
            return {"status": response.status_code, "data": result}
        except:
            return {"status": response.status_code, "data": None}

    def test_authentication_flow(self):
        print("=== Testing Authentication Flow ===")

        # Register
        register_data = {
            "email": f"test{int(time.time())}@example.com",
            "password": "testpassword123",
            "first_name": "Test",
            "last_name": "User"
        }

        register_result = self.request("POST", "/auth/register", register_data)
        if register_result["status"] == 201:
            self.access_token = register_result["data"]["data"]["access_token"]
            self.refresh_token = register_result["data"]["data"]["refresh_token"]
            self.user_id = register_result["data"]["data"]["user"]["id"]
            print("✅ Registration successful")
        else:
            print("❌ Registration failed")
            return False

        # Login
        login_data = {
            "email": register_data["email"],
            "password": register_data["password"]
        }

        login_result = self.request("POST", "/auth/login", login_data)
        if login_result["status"] == 200:
            self.access_token = login_result["data"]["data"]["access_token"]
            print("✅ Login successful")
        else:
            print("❌ Login failed")

        return True

    def test_user_management(self):
        print("=== Testing User Management ===")

        # Get user profile
        profile_result = self.request("GET", f"/users/{self.user_id}", requires_auth=True)
        if profile_result["status"] == 200:
            print("✅ Get user profile successful")
        else:
            print("❌ Get user profile failed")

        # Update user profile
        update_data = {
            "first_name": "Updated Test",
            "last_name": "Updated User"
        }

        update_result = self.request("PUT", f"/users/{self.user_id}", update_data, requires_auth=True)
        if update_result["status"] == 200:
            print("✅ Update user profile successful")
        else:
            print("❌ Update user profile failed")

        # List users
        list_result = self.request("GET", "/users?page=1&limit=10", requires_auth=True)
        if list_result["status"] == 200:
            print("✅ List users successful")
        else:
            print("❌ List users failed")

    def run_all_tests(self):
        if self.test_authentication_flow():
            self.test_user_management()
        print("=== All Tests Completed ===")

# Run the tests
if __name__ == "__main__":
    tester = APITester()
    tester.run_all_tests()
```

## Test Scenarios

### 1. Happy Path Testing

- User registration → Login → Access protected resources → Logout
- Token refresh flow
- CRUD operations on user profiles
- Pagination and filtering

### 2. Error Handling Testing

- Invalid credentials
- Expired tokens
- Missing authorization headers
- Invalid request bodies
- Non-existent resources
- Validation errors

### 3. Security Testing

- Access protected endpoints without authentication
- Use expired tokens
- Use invalid tokens
- SQL injection attempts (should be prevented by GORM)
- XSS attempts in request bodies

### 4. Performance Testing

- Concurrent user registrations
- High-frequency token refresh
- Large pagination requests
- Search with various query patterns

## Automated Testing Setup

### GitHub Actions Example

```yaml
name: API Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:13
        env:
          POSTGRES_PASSWORD: testpassword
          POSTGRES_DB: testdb
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v2

      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.19

      - name: Run migrations
        run: go run cmd/migrate/main.go up
        env:
          DATABASE_URL: postgres://postgres:testpassword@localhost:5432/testdb?sslmode=disable

      - name: Start server
        run: go run cmd/server/main.go &
        env:
          DATABASE_URL: postgres://postgres:testpassword@localhost:5432/testdb?sslmode=disable
          JWT_SECRET: test-secret
          PORT: 8080

      - name: Wait for server
        run: sleep 5

      - name: Run API tests
        run: node test-api.js
```

## Troubleshooting

### Common Issues

1. **Server not running**: Ensure the server is started with `go run cmd/server/main.go`
2. **Database connection**: Check PostgreSQL is running and environment variables are set
3. **Port conflicts**: Make sure port 8080 is available
4. **CORS issues**: The server includes CORS middleware for cross-origin requests
5. **Token expiration**: Access tokens expire in 15 minutes, use refresh tokens

### Debug Tips

1. **Check server logs** for detailed error information
2. **Use Swagger UI** for interactive testing and debugging
3. **Verify environment variables** are properly set
4. **Check database migrations** are applied
5. **Test with simple cURL commands** first before complex scenarios

This comprehensive testing guide should help you thoroughly test all API endpoints and ensure the Go Fiber Template API works correctly in various scenarios.
