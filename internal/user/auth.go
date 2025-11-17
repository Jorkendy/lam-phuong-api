package user

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"lam-phuong-api/internal/response"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.InvalidToken(c, "Invalid authorization header format. Use: Bearer <token>")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := ValidateToken(tokenString, jwtSecret)
		if err != nil {
			// Check if token is expired (jwt/v5 returns errors with "expired" in the message)
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "expired") || strings.Contains(errMsg, "exp") {
				response.ExpiredToken(c)
			} else {
				response.InvalidToken(c, "Invalid or expired token")
			}
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Login handles user authentication
func (h *Handler) Login(c *gin.Context, jwtSecret string, tokenExpiry time.Duration) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request data", map[string]interface{}{
			"validation_error": err.Error(),
		})
		return
	}

	// Get user by email
	user, ok := h.repo.GetByEmail(req.Email)
	if !ok {
		response.InvalidAuth(c, "Invalid email or password")
		return
	}

	// Check user status
	if user.Status == StatusDisabled {
		response.Error(c, http.StatusForbidden, response.ErrCodeForbidden, "Your account has been disabled. Please contact support.", nil)
		return
	}

	// Verify password
	if !CheckPassword(user.Password, req.Password) {
		response.InvalidAuth(c, "Invalid email or password")
		return
	}

	// Generate JWT token
	token, err := GenerateToken(user, jwtSecret, tokenExpiry)
	if err != nil {
		response.InternalError(c, "Failed to generate token")
		return
	}

	// Remove password from user object
	user.Password = ""

	// Return token response
	tokenResp := TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(tokenExpiry.Seconds()),
		User:        user,
	}
	response.Success(c, http.StatusOK, tokenResp, "Login successful")
}
