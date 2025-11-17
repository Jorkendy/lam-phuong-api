package user

import (
	"lam-phuong-api/internal/response"

	"github.com/gin-gonic/gin"
)

// RequireRole creates a middleware that requires the user to have one of the specified roles
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (set by AuthMiddleware)
		userRole, exists := c.Get("user_role")
		if !exists {
			response.Forbidden(c, "User role not found in context")
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			response.Forbidden(c, "Invalid user role type")
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
			response.Forbidden(c, "Insufficient permissions. Required roles: "+joinRoles(allowedRoles))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin is a convenience middleware that requires admin role or higher (Super Admin)
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(RoleAdmin, RoleSuperAdmin)
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
