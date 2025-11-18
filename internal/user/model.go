package user

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Airtable field names
const (
	FieldEmail                  = "Email"
	FieldPassword               = "Password"
	FieldRole                   = "Role"
	FieldStatus                 = "Status"
	FieldEmailVerificationToken = "Email Verification Token"
	FieldCreatedAt              = "Created At"
	FieldUpdatedAt              = "Updated At"
)

// User status constants
const (
	StatusPending  = "Pending"
	StatusActive   = "Active"
	StatusDisabled = "Disabled"
)

// User roles
const (
	RoleSuperAdmin = "Super Admin"
	RoleAdmin      = "Admin"
	RoleUser       = "User"
)

// ValidRoles contains all valid user roles
var ValidRoles = []string{RoleSuperAdmin, RoleAdmin, RoleUser}

// Helper functions
func getStringField(fields map[string]interface{}, key string) string {
	if val, ok := fields[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// User represents a user in the system
type User struct {
	ID                    string `json:"id"`
	Email                 string `json:"email"`
	Password              string `json:"-"` // Never serialize password in JSON responses
	Role                  string `json:"role"`
	Status                string `json:"status"` // "pending", "active", or "disabled"
	EmailVerificationToken string `json:"-"` // Never serialize verification token in JSON responses
}

// ToAirtableFields converts a User to Airtable fields format (for creation)
// Deprecated: Use ToAirtableFieldsForCreate() instead
func (u *User) ToAirtableFields() map[string]interface{} {
	return u.ToAirtableFieldsForCreate()
}

// HashPassword hashes a plain text password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword compares a plain text password with a hashed password
func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// TokenResponse represents the response after successful authentication
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	User        User   `json:"user"`
}

// TokenResponseWrapper wraps TokenResponse in the standard API response format for Swagger
// @Description Response containing authentication token
type TokenResponseWrapper struct {
	Success bool          `json:"success" example:"true"`
	Data    TokenResponse `json:"data"`
	Message string        `json:"message" example:"Login successful"`
}

// UserResponseWrapper wraps User in the standard API response format for Swagger
// @Description Response containing a single user
type UserResponseWrapper struct {
	Success bool   `json:"success" example:"true"`
	Data    User   `json:"data"`
	Message string `json:"message" example:"User retrieved successfully"`
}

// UsersResponseWrapper wraps array of Users in the standard API response format for Swagger
// @Description Response containing a list of users
type UsersResponseWrapper struct {
	Success bool   `json:"success" example:"true"`
	Data    []User `json:"data"`
	Message string `json:"message" example:"Users retrieved successfully"`
}

// GenerateToken generates a JWT token for the user
func GenerateToken(user User, secretKey string, expiresIn time.Duration) (string, error) {
	expirationTime := time.Now().Add(expiresIn)
	role := user.Role
	if role == "" {
		role = RoleUser // Default role
	}
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString, secretKey string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

// GenerateVerificationToken generates a secure random token for email verification
func GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate verification token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
