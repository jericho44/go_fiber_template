# Swagger UI Testing Guide

This comprehensive guide provides step-by-step instructions for testing the Go Fiber Template API using the interactive Swagger UI interface.

## Prerequisites

1. **Server Running**: Ensure the API server is running on `http://localhost:8080`
2. **Database**: PostgreSQL database should be set up and migrations applied
3. **Environment**: Proper environment variables configured (see `.env.example`)

## Accessing Swagger UI

Open your web browser and navigate to:

```
http://localhost:8080/swagger/
```

You should see the Swagger UI interface with the Go Fiber Template API documentation.

## Understanding the Swagger UI Interface

### Main Components

1. **API Title and Description**: At the top, you'll see "Go Fiber Template API v1.0"
2. **Authentication Section**: Detailed instructions on how to authenticate
3. **Authorize Button**: Green "Authorize" button in the top-right for setting up authentication
4. **Endpoint Groups**:
   - **Authentication**: User registration, login, logout, and token management
   - **Users**: User management operations (requires authentication)

### Response Format Information

The documentation explains that all API responses follow this structure:

```json
{
  "success": true,
  "message": "Operation successful",
  "data": {...},
  "error": null,
  "meta": {...}
}
```

## Step-by-Step Testing Guide

### Phase 1: Authentication Flow Testing

#### Step 1: Register a New User

1. **Locate the Authentication Section**

   - Scroll down to find the "Authentication" section
   - Click to expand it if it's collapsed

2. **Open the Register Endpoint**

   - Find `POST /auth/register`
   - Click on it to expand the endpoint details

3. **Test User Registration**

   - Click the **"Try it out"** button
   - You'll see an editable request body with example data
   - Modify the request body with your test data:

   ```json
   {
     "email": "testuser@example.com",
     "password": "testpassword123",
     "first_name": "Test",
     "last_name": "User"
   }
   ```

   - Click **"Execute"**

4. **Review the Response**
   - Check the response status (should be 201 for success)
   - In the response body, you should see:
     - `success: true`
     - User information in the `data.user` object
     - `access_token` and `refresh_token` in the `data` object
   - **Important**: Copy the `access_token` value - you'll need it for authentication

#### Step 2: Set Up Authentication

1. **Click the Authorize Button**

   - Look for the green "Authorize" button at the top of the page
   - Click it to open the authorization dialog

2. **Enter Your Access Token**

   - In the "Value" field, enter: `Bearer {your-access-token}`
   - Replace `{your-access-token}` with the actual token from the registration response
   - Example: `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`
   - Click **"Authorize"**
   - Click **"Close"**

3. **Verify Authentication Setup**
   - You should now see a lock icon (🔒) next to protected endpoints
   - The "Authorize" button should show as "Logout"

#### Step 3: Test Login Endpoint

1. **Open the Login Endpoint**

   - Find `POST /auth/login`
   - Click to expand it

2. **Test User Login**

   - Click **"Try it out"**
   - Enter the same credentials you used for registration:

   ```json
   {
     "email": "testuser@example.com",
     "password": "testpassword123"
   }
   ```

   - Click **"Execute"**

3. **Verify Login Response**
   - Should return status 200
   - Response should contain new `access_token` and `refresh_token`
   - User information should be included in the response

#### Step 4: Test Token Refresh

1. **Open the Refresh Token Endpoint**

   - Find `POST /auth/refresh`
   - Click to expand it

2. **Test Token Refresh**

   - Click **"Try it out"**
   - Enter the refresh token from your login/registration response:

   ```json
   {
     "refresh_token": "your_refresh_token_here"
   }
   ```

   - Click **"Execute"**

3. **Verify Refresh Response**
   - Should return status 200
   - Response should contain a new `access_token`
   - Note the new `expires_at` timestamp

### Phase 2: User Management Testing

Now that you're authenticated, you can test the protected user endpoints.

#### Step 5: List Users

1. **Open the Users Section**

   - Scroll down to the "Users" section
   - Notice the lock icons (🔒) indicating these endpoints require authentication

2. **Test List Users Endpoint**

   - Find `GET /users`
   - Click to expand it
   - Click **"Try it out"**

3. **Test with Default Parameters**

   - Leave all query parameters empty for default behavior
   - Click **"Execute"**

4. **Test with Custom Parameters**

   - Try different combinations:
     - `page`: 1
     - `limit`: 5
     - `search`: "test" (to search for your test user)
     - `is_active`: true
     - `sort_by`: "created_at"
     - `sort_order`: "desc"
   - Click **"Execute"** after each change

5. **Review Pagination Response**
   - Check the `meta` object in the response:
   ```json
   {
     "meta": {
       "page": 1,
       "limit": 10,
       "total": 1,
       "total_pages": 1
     }
   }
   ```

#### Step 6: Get User Profile

1. **Open Get User Profile Endpoint**

   - Find `GET /users/{id}`
   - Click to expand it

2. **Test Getting User Profile**

   - Click **"Try it out"**
   - Enter a user ID (use the ID from your registered user, typically `1`)
   - Click **"Execute"**

3. **Verify Profile Response**
   - Should return status 200
   - Response should contain complete user information
   - Password should not be visible (excluded from JSON response)

#### Step 7: Update User Profile

1. **Open Update User Profile Endpoint**

   - Find `PUT /users/{id}`
   - Click to expand it

2. **Test Profile Update**

   - Click **"Try it out"**
   - Enter the user ID
   - Modify the request body with new information:

   ```json
   {
     "first_name": "Updated Test",
     "last_name": "Updated User",
     "is_active": true
   }
   ```

   - Click **"Execute"**

3. **Verify Update Response**
   - Should return status 200
   - Response should show the updated user information
   - `updated_at` timestamp should be more recent

#### Step 8: Test User Deletion

1. **Open Delete User Endpoint**

   - Find `DELETE /users/{id}`
   - Click to expand it

2. **Test User Deletion**

   - Click **"Try it out"**
   - Enter a user ID (you might want to create another user first)
   - Click **"Execute"**

3. **Verify Deletion Response**
   - Should return status 200
   - Response should confirm successful deletion
   - Try to get the user profile again - should return 404

### Phase 3: Error Handling Testing

#### Step 9: Test Authentication Errors

1. **Test Invalid Login**

   - Go back to `POST /auth/login`
   - Try with invalid credentials:

   ```json
   {
     "email": "invalid@example.com",
     "password": "wrongpassword"
   }
   ```

   - Should return status 401 with error details

2. **Test Validation Errors**
   - Try registration with invalid data:
   ```json
   {
     "email": "invalid-email",
     "password": "123",
     "first_name": "",
     "last_name": ""
   }
   ```
   - Should return status 400 with validation error details

#### Step 10: Test Unauthorized Access

1. **Clear Authorization**

   - Click the "Logout" button (previously "Authorize")
   - This removes your authentication token

2. **Try Protected Endpoints**

   - Attempt to access `GET /users`
   - Should return status 401 (Unauthorized)

3. **Re-authenticate**
   - Click "Authorize" again
   - Enter your Bearer token to continue testing

#### Step 11: Test Logout

1. **Open Logout Endpoint**

   - Find `POST /auth/logout`
   - Click to expand it

2. **Test Logout**

   - Click **"Try it out"**
   - Optionally include your refresh token in the request body:

   ```json
   {
     "refresh_token": "your_refresh_token_here"
   }
   ```

   - Click **"Execute"**

3. **Verify Logout**
   - Should return status 200
   - Try using the same access token for other requests - should now return 401

## Advanced Testing Scenarios

### Testing Edge Cases

1. **Large Pagination Requests**

   - Test `GET /users` with `limit=100` (maximum allowed)
   - Test with `limit=101` (should be rejected)

2. **Invalid User IDs**

   - Test `GET /users/999999` (non-existent user)
   - Test `GET /users/abc` (invalid ID format)

3. **Expired Token Handling**
   - Wait for your access token to expire (15 minutes)
   - Try accessing protected endpoints
   - Use refresh token to get new access token

### Performance Testing

1. **Concurrent Requests**

   - Open multiple browser tabs with Swagger UI
   - Execute the same endpoint simultaneously
   - Observe response times and behavior

2. **Search Performance**
   - Test `GET /users` with various search terms
   - Try empty search, single characters, and long strings

## Troubleshooting Common Issues

### Authentication Issues

**Problem**: "Unauthorized" errors even with valid token
**Solution**:

- Ensure token format is `Bearer {token}` (with space)
- Check if token has expired (15-minute lifetime)
- Verify you clicked "Authorize" after entering the token

**Problem**: "Invalid authorization header" error
**Solution**:

- Make sure you're including "Bearer " prefix
- Check for extra spaces or characters in the token

### Request Issues

**Problem**: Validation errors on seemingly valid data
**Solution**:

- Check required field constraints (email format, password length, etc.)
- Ensure JSON format is correct (no trailing commas, proper quotes)

**Problem**: 404 errors on existing endpoints
**Solution**:

- Verify the server is running on the correct port (8080)
- Check that you're using the correct base path (/api/v1)

### Response Issues

**Problem**: Unexpected response format
**Solution**:

- All responses follow the standard format with `success`, `message`, `data`, `error`, and `meta` fields
- Check the `error` object for detailed error information

## Best Practices for API Testing

1. **Test in Order**: Follow the authentication flow before testing protected endpoints
2. **Save Tokens**: Keep track of your access and refresh tokens during testing
3. **Check Status Codes**: Always verify the HTTP status code matches expectations
4. **Read Error Messages**: Error responses include detailed information in the `error` object
5. **Test Edge Cases**: Try invalid inputs, missing fields, and boundary conditions
6. **Use Real Data**: Test with realistic data that matches your use case
7. **Clean Up**: Delete test users when finished to keep the database clean

## API Response Reference

### Success Response Format

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
  }
}
```

### Error Response Format

```json
{
  "success": false,
  "message": "Request failed",
  "data": null,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": {
      // Field-specific validation errors
    }
  },
  "meta": null
}
```

### Common Error Codes

- `VALIDATION_ERROR`: Request validation failed
- `UNAUTHORIZED`: Invalid credentials or expired token
- `NOT_FOUND`: Resource not found
- `CONFLICT`: Resource already exists (e.g., email already registered)
- `INTERNAL_ERROR`: Server error occurred

## Conclusion

The Swagger UI provides a powerful interface for testing and exploring the Go Fiber Template API. By following this guide, you can:

- Understand the complete authentication flow
- Test all available endpoints interactively
- Verify error handling and edge cases
- Learn the API structure and response formats

This interactive documentation serves as both a testing tool and a reference for developers integrating with the API.
