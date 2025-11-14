# Authorization Guide

This guide explains how to use role-based authorization in the Lam Phuong API.

## Overview

The API uses JWT-based authentication with role-based access control (RBAC). Users are assigned roles that determine what actions they can perform.

## Available Roles

The API supports three roles:

- **Super Admin** (`"Super Admin"`) - Highest level of access
- **Admin** (`"Admin"`) - Administrative access
- **User** (`"User"`) - Standard user access (default)

## How Authorization Works

1. **Authentication**: Users login via `/api/auth/login` and receive a JWT token
2. **Token Validation**: The `AuthMiddleware` validates the JWT token and extracts user information
3. **Role Checking**: The `RequireRole` middleware checks if the user has the required role(s)
4. **Context Access**: User information (ID, email, role) is available in the Gin context

## Using Authorization Middleware

### 1. Require Admin Role

Use `RequireAdmin()` to restrict access to admin users only:

```go
adminRoutes := router.Group("")
adminRoutes.Use(user.RequireAdmin())
{
    adminRoutes.GET("/admin-only", adminHandler)
    adminRoutes.POST("/admin-only", adminHandler)
}
```

### 2. Require Specific Role(s)

Use `RequireRole()` to allow multiple roles:

```go
// Allow both Admin and Super Admin
superAdminRoutes := router.Group("")
superAdminRoutes.Use(user.RequireRole(user.RoleSuperAdmin, user.RoleAdmin))
{
    superAdminRoutes.GET("/super-admin-only", superAdminHandler)
}
```

### 3. Require Any Authenticated User

If you only need authentication (any role), just use `AuthMiddleware`:

```go
protected := router.Group("")
protected.Use(user.AuthMiddleware(jwtSecret))
{
    // Any authenticated user can access these routes
    protected.GET("/profile", getProfileHandler)
}
```

## Examples

### Example 1: Admin-Only Route

```go
// In router.go
func NewRouter(...) *gin.Engine {
    router := gin.Default()
    
    api := router.Group("/api")
    protected := api.Group("")
    protected.Use(user.AuthMiddleware(jwtSecret))
    {
        // Admin-only routes
        adminRoutes := protected.Group("")
        adminRoutes.Use(user.RequireAdmin())
        {
            adminRoutes.DELETE("/users/:id", deleteUserHandler)
        }
    }
    
    return router
}
```

### Example 2: Multiple Roles

```go
// Allow both Admin and Super Admin
managementRoutes := protected.Group("")
managementRoutes.Use(user.RequireRole(user.RoleSuperAdmin, user.RoleAdmin))
{
    managementRoutes.GET("/reports", getReportsHandler)
    managementRoutes.POST("/settings", updateSettingsHandler)
}
```

### Example 3: Accessing User Information in Handlers

You can access user information from the Gin context:

```go
func MyHandler(c *gin.Context) {
    // Get user ID
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
        return
    }
    
    // Get user email
    userEmail, _ := c.Get("user_email")
    
    // Get user role
    userRole, _ := c.Get("user_role")
    
    // Use the information
    c.JSON(http.StatusOK, gin.H{
        "user_id": userID,
        "email": userEmail,
        "role": userRole,
    })
}
```

### Example 4: Conditional Logic Based on Role

```go
func ConditionalHandler(c *gin.Context) {
    userRole, exists := c.Get("user_role")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
        return
    }
    
    role := userRole.(string)
    
    // Different behavior based on role
    if role == user.RoleSuperAdmin {
        // Super admin can see everything
        data := getAllData()
        c.JSON(http.StatusOK, data)
    } else if role == user.RoleAdmin {
        // Admin can see limited data
        data := getLimitedData()
        c.JSON(http.StatusOK, data)
    } else {
        // Regular users see only their own data
        userID, _ := c.Get("user_id")
        data := getUserData(userID.(string))
        c.JSON(http.StatusOK, data)
    }
}
```

## Current Implementation

### User Management Routes

Currently, all user management routes require admin role:

```go
// In internal/server/router.go
adminRoutes := protected.Group("")
adminRoutes.Use(user.RequireAdmin())
{
    adminRoutes.GET("/users", userHandler.ListUsers)
    adminRoutes.POST("/users", userHandler.CreateUser)
    adminRoutes.DELETE("/users/:id", userHandler.DeleteUser)
}
```

### Location Routes

Location routes require authentication but any role can access:

```go
// In internal/server/router.go
protected := api.Group("")
protected.Use(user.AuthMiddleware(jwtSecret))
{
    locationHandler.RegisterRoutes(protected)
}
```

## Adding Authorization to New Routes

### Step 1: Create Your Handler

```go
func NewMyHandler(repo MyRepository) *MyHandler {
    return &MyHandler{repo: repo}
}

func (h *MyHandler) MyProtectedHandler(c *gin.Context) {
    // Your handler logic
}
```

### Step 2: Register Routes with Authorization

```go
func (h *MyHandler) RegisterRoutes(router *gin.RouterGroup) {
    // Public route (no auth required)
    router.GET("/public", h.PublicHandler)
    
    // Protected route (auth required, any role)
    protected := router.Group("")
    protected.Use(user.AuthMiddleware(jwtSecret))
    {
        protected.GET("/protected", h.ProtectedHandler)
        
        // Admin-only route
        adminRoutes := protected.Group("")
        adminRoutes.Use(user.RequireAdmin())
        {
            adminRoutes.GET("/admin-only", h.AdminOnlyHandler)
        }
    }
}
```

### Step 3: Use in Main Router

```go
// In internal/server/router.go
func NewRouter(...) *gin.Engine {
    // ... existing code ...
    
    protected := api.Group("")
    protected.Use(user.AuthMiddleware(jwtSecret))
    {
        myHandler.RegisterRoutes(protected)
    }
}
```

## Testing Authorization

### 1. Login as Admin

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "password123"
  }'
```

Response:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "user": {
    "id": "1",
    "email": "admin@example.com",
    "role": "Admin"
  }
}
```

### 2. Access Admin-Only Route

```bash
curl -X GET http://localhost:8080/api/users \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 3. Test Unauthorized Access

If a user with role "User" tries to access an admin-only route:

```bash
# Login as regular user
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'

# Try to access admin route (will fail with 403)
curl -X GET http://localhost:8080/api/users \
  -H "Authorization: Bearer <user_token>"
```

Response:
```json
{
  "error": "Insufficient permissions. Required roles: Admin"
}
```

## Best Practices

1. **Always use middleware for route protection** - Don't manually check roles in handlers unless you need conditional logic
2. **Use the most restrictive role** - Start with the most restrictive access and relax as needed
3. **Validate user input** - Even with authorization, always validate and sanitize user input
4. **Log authorization failures** - Consider logging failed authorization attempts for security monitoring
5. **Use context for user info** - Always get user information from Gin context, not from request parameters

## Troubleshooting

### Issue: "User role not found in context"

**Cause**: The route is not protected by `AuthMiddleware` or the middleware is not applied correctly.

**Solution**: Ensure `AuthMiddleware` is applied before `RequireRole`:

```go
protected := router.Group("")
protected.Use(user.AuthMiddleware(jwtSecret))  // Must be first
{
    adminRoutes := protected.Group("")
    adminRoutes.Use(user.RequireAdmin())  // Then role check
    {
        // routes
    }
}
```

### Issue: "Invalid or expired token"

**Cause**: The JWT token is invalid, expired, or missing.

**Solution**: 
- Check that the token is included in the Authorization header: `Authorization: Bearer <token>`
- Verify the token hasn't expired
- Ensure the JWT secret matches between token generation and validation

### Issue: Always getting 403 Forbidden

**Cause**: The user's role doesn't match the required role(s).

**Solution**:
- Check the user's role in the database/Airtable
- Verify the role constant matches exactly (case-sensitive): `"Admin"` not `"admin"`
- Ensure the role is included in the JWT token claims

## Summary

- Use `AuthMiddleware` for authentication (any role)
- Use `RequireAdmin()` for admin-only routes
- Use `RequireRole(...)` for custom role requirements
- Access user info from Gin context: `c.Get("user_id")`, `c.Get("user_email")`, `c.Get("user_role")`
- Always apply `AuthMiddleware` before role-based middleware

