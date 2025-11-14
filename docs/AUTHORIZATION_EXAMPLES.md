# Authorization Examples

Quick reference examples for using authorization in the Lam Phuong API.

## Quick Reference

### Available Middleware

```go
// Authentication only (any role)
user.AuthMiddleware(jwtSecret)

// Admin only
user.RequireAdmin()

// Specific role(s)
user.RequireRole(user.RoleSuperAdmin, user.RoleAdmin)

// Any authenticated user (Admin or User)
user.RequireAnyRole()  // Equivalent to RequireRole(RoleAdmin, RoleUser)
```

### Available Roles

```go
user.RoleSuperAdmin  // "Super Admin"
user.RoleAdmin       // "Admin"
user.RoleUser        // "User" (default)
```

## Example 1: Simple Admin Route

```go
// In router.go
func NewRouter(...) *gin.Engine {
    router := gin.Default()
    
    api := router.Group("/api")
    protected := api.Group("")
    protected.Use(user.AuthMiddleware(jwtSecret))
    {
        // Admin-only route
        adminRoutes := protected.Group("")
        adminRoutes.Use(user.RequireAdmin())
        {
            adminRoutes.GET("/admin/dashboard", getDashboardHandler)
        }
    }
    
    return router
}
```

## Example 2: Multiple Role Levels

```go
protected := api.Group("")
protected.Use(user.AuthMiddleware(jwtSecret))
{
    // Super Admin only
    superAdminRoutes := protected.Group("")
    superAdminRoutes.Use(user.RequireRole(user.RoleSuperAdmin))
    {
        superAdminRoutes.GET("/system/settings", getSystemSettingsHandler)
        superAdminRoutes.POST("/system/settings", updateSystemSettingsHandler)
    }
    
    // Admin or Super Admin
    adminRoutes := protected.Group("")
    adminRoutes.Use(user.RequireRole(user.RoleSuperAdmin, user.RoleAdmin))
    {
        adminRoutes.GET("/reports", getReportsHandler)
    }
    
    // Any authenticated user
    protected.GET("/profile", getProfileHandler)
}
```

## Example 3: Handler with User Context

```go
func GetProfileHandler(c *gin.Context) {
    // Get user ID from context
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
        return
    }
    
    // Get user email
    userEmail, _ := c.Get("user_email")
    
    // Get user role
    userRole, _ := c.Get("user_role")
    
    c.JSON(http.StatusOK, gin.H{
        "id":    userID,
        "email": userEmail,
        "role":  userRole,
    })
}
```

## Example 4: Conditional Logic Based on Role

```go
func GetDataHandler(c *gin.Context) {
    userRole, _ := c.Get("user_role")
    role := userRole.(string)
    
    var data interface{}
    
    switch role {
    case user.RoleSuperAdmin:
        // Super admin sees everything
        data = getAllData()
    case user.RoleAdmin:
        // Admin sees department data
        data = getDepartmentData()
    default:
        // Regular users see only their own data
        userID, _ := c.Get("user_id")
        data = getUserData(userID.(string))
    }
    
    c.JSON(http.StatusOK, data)
}
```

## Example 5: Creating a New Protected Module

```go
// internal/report/handler.go
package report

import (
    "net/http"
    "lam-phuong-api/internal/user"
    "github.com/gin-gonic/gin"
)

type Handler struct {
    repo Repository
}

func NewHandler(repo Repository) *Handler {
    return &Handler{repo: repo}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup, jwtSecret string) {
    // Public route
    router.GET("/reports/public", h.GetPublicReport)
    
    // Protected routes
    protected := router.Group("")
    protected.Use(user.AuthMiddleware(jwtSecret))
    {
        // Any authenticated user
        protected.GET("/reports/user", h.GetUserReport)
        
        // Admin only
        adminRoutes := protected.Group("")
        adminRoutes.Use(user.RequireAdmin())
        {
            adminRoutes.GET("/reports/admin", h.GetAdminReport)
            adminRoutes.GET("/reports/all", h.GetAllReports)
        }
    }
}
```

## Example 6: Testing with cURL

### 1. Login and get token

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password123"}' \
  | jq -r '.access_token')

echo "Token: $TOKEN"
```

### 2. Access protected route

```bash
curl -X GET http://localhost:8080/api/users \
  -H "Authorization: Bearer $TOKEN"
```

### 3. Test unauthorized access

```bash
# Login as regular user
USER_TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}' \
  | jq -r '.access_token')

# Try to access admin route (will fail with 403)
curl -X GET http://localhost:8080/api/users \
  -H "Authorization: Bearer $USER_TOKEN"
```

## Example 7: Complete Route Setup

```go
// internal/server/router.go
func NewRouter(
    locationHandler *location.Handler,
    userHandler *user.Handler,
    reportHandler *report.Handler,
    jwtSecret string,
) *gin.Engine {
    router := gin.Default()
    
    // CORS middleware
    router.Use(cors.New(cors.Config{
        AllowOrigins: []string{"*"},
        AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
    }))
    
    api := router.Group("/api")
    {
        // Public routes
        api.GET("/ping", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{"status": "ok"})
        })
        
        // Auth routes (public)
        userHandler.RegisterRoutes(api)
        
        // Protected routes
        protected := api.Group("")
        protected.Use(user.AuthMiddleware(jwtSecret))
        {
            // Location routes (any authenticated user)
            locationHandler.RegisterRoutes(protected)
            
            // Report routes (any authenticated user)
            reportHandler.RegisterRoutes(protected, jwtSecret)
            
            // User management (admin only)
            adminRoutes := protected.Group("")
            adminRoutes.Use(user.RequireAdmin())
            {
                adminRoutes.GET("/users", userHandler.ListUsers)
                adminRoutes.POST("/users", userHandler.CreateUser)
                adminRoutes.DELETE("/users/:id", userHandler.DeleteUser)
            }
            
            // System settings (super admin only)
            superAdminRoutes := protected.Group("")
            superAdminRoutes.Use(user.RequireRole(user.RoleSuperAdmin))
            {
                superAdminRoutes.GET("/system/config", getSystemConfigHandler)
                superAdminRoutes.POST("/system/config", updateSystemConfigHandler)
            }
        }
    }
    
    // Swagger
    router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    
    return router
}
```

## Common Patterns

### Pattern 1: Nested Groups

```go
protected := api.Group("")
protected.Use(user.AuthMiddleware(jwtSecret))
{
    // Level 1: Admin routes
    adminRoutes := protected.Group("")
    adminRoutes.Use(user.RequireAdmin())
    {
        // Level 2: Super Admin routes (nested)
        superAdminRoutes := adminRoutes.Group("")
        superAdminRoutes.Use(user.RequireRole(user.RoleSuperAdmin))
        {
            superAdminRoutes.GET("/super-admin-only", handler)
        }
        
        // Level 2: Regular admin routes
        adminRoutes.GET("/admin-only", handler)
    }
}
```

### Pattern 2: Separate Route Groups

```go
protected := api.Group("")
protected.Use(user.AuthMiddleware(jwtSecret))
{
    // Group 1: Admin routes
    adminGroup := protected.Group("/admin")
    adminGroup.Use(user.RequireAdmin())
    {
        adminGroup.GET("/dashboard", getDashboardHandler)
    }
    
    // Group 2: Super Admin routes
    superAdminGroup := protected.Group("/super-admin")
    superAdminGroup.Use(user.RequireRole(user.RoleSuperAdmin))
    {
        superAdminGroup.GET("/settings", getSettingsHandler)
    }
}
```

### Pattern 3: Inline Middleware

```go
protected := api.Group("")
protected.Use(
    user.AuthMiddleware(jwtSecret),
    user.RequireAdmin(),  // Applied to all routes in this group
)
{
    protected.GET("/admin-only", handler)
}
```

## Error Responses

### 401 Unauthorized (No token or invalid token)

```json
{
  "error": "Authorization header required"
}
```

or

```json
{
  "error": "Invalid or expired token"
}
```

### 403 Forbidden (Wrong role)

```json
{
  "error": "Insufficient permissions. Required roles: Admin"
}
```

## Tips

1. **Always apply AuthMiddleware first** before role-based middleware
2. **Use nested groups** for hierarchical permissions
3. **Access user info from context** - don't trust client-provided user IDs
4. **Validate roles** when creating users to prevent invalid roles
5. **Log authorization failures** for security monitoring

