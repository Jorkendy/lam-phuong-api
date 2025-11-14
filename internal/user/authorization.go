package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole creates a middleware that requires the user to have one of the specified roles
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (set by AuthMiddleware)
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "User role not found in context"})
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid user role type"})
			c.Abort()
			return
		}

		// Check if user's role is in the allowed roles list
		hasPermission := false
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions. Required roles: " + joinRoles(allowedRoles),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin is a convenience middleware that requires admin role
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(RoleAdmin)
}

// RequireAnyRole is a convenience middleware that requires user or admin role
func RequireAnyRole() gin.HandlerFunc {
	return RequireRole(RoleAdmin, RoleUser)
}

// joinRoles joins roles into a comma-separated string
func joinRoles(roles []string) string {
	if len(roles) == 0 {
		return ""
	}
	result := roles[0]
	for i := 1; i < len(roles); i++ {
		result += ", " + roles[i]
	}
	return result
}

