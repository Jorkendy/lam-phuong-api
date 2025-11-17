package user

import (
	"net/http"
	"strings"
	"time"

	"lam-phuong-api/internal/email"
	"lam-phuong-api/internal/response"

	"github.com/gin-gonic/gin"
)

// Handler exposes HTTP handlers for the user resource
type Handler struct {
	repo         Repository
	jwtSecret    string
	tokenExpiry  time.Duration
	emailService *email.Service
	baseURL      string
}

// NewHandler creates a handler with the provided repository
func NewHandler(repo Repository, jwtSecret string, tokenExpiry time.Duration, emailService *email.Service, baseURL string) *Handler {
	return &Handler{
		repo:         repo,
		jwtSecret:    jwtSecret,
		tokenExpiry:  tokenExpiry,
		emailService: emailService,
		baseURL:      baseURL,
	}
}

// RegisterRoutes attaches user routes to the supplied router group
// Only registers public auth routes. Protected routes should be registered separately in router.go
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// Public routes only
	router.POST("/auth/register", h.RegisterHandler)
	router.POST("/auth/login", h.LoginHandler)
	router.GET("/auth/verify-email", h.VerifyEmailHandler)
}

// Register godoc
// @Summary      User registration
// @Description  Register a new user account with email and password. A verification email will be sent to the provided email address.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      RegisterRequest  true  "Registration credentials"
// @Success      201         {object}  response.Response  "User registered successfully. Please check your email to verify your account."
// @Failure      400         {object}  response.ErrorResponse  "Validation error"
// @Failure      409         {object}  response.ErrorResponse  "Email already registered"
// @Failure      500         {object}  response.ErrorResponse  "Internal server error"
// @Router       /auth/register [post]
func (h *Handler) RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request data", map[string]interface{}{
			"validation_error": err.Error(),
		})
		return
	}

	// Check if user already exists
	_, exists := h.repo.GetByEmail(req.Email)
	if exists {
		response.DuplicateEmail(c)
		return
	}

	// Hash the password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		response.InternalError(c, "Failed to hash password")
		return
	}

	// Generate verification token
	verificationToken, err := email.GenerateVerificationToken()
	if err != nil {
		response.InternalError(c, "Failed to generate verification token")
		return
	}

	// Create user with default "User" role and pending status
	user := User{
		Email:                  req.Email,
		Password:               hashedPassword,
		Role:                   RoleUser, // Always "User" role for public registration
		Status:                 StatusPending,
		EmailVerificationToken: verificationToken,
	}

	// Create in repository (repository handles Airtable sync if configured)
	created, err := h.repo.Create(c.Request.Context(), user)
	if err != nil {
		// Check if it's a duplicate email error (race condition)
		if strings.Contains(err.Error(), "already exists") {
			response.DuplicateEmail(c)
			return
		}
		response.InternalError(c, "Failed to create user: "+err.Error())
		return
	}

	// Send verification email
	if h.emailService != nil {
		if err := h.emailService.SendVerificationEmail(created.Email, verificationToken, h.baseURL); err != nil {
			// Log error but don't fail registration - email can be resent later
			c.Error(err) // Log to gin context for error tracking
		}
	}

	// Remove sensitive fields from response
	created.Password = ""
	created.EmailVerificationToken = ""

	response.Success(c, http.StatusCreated, created, "User registered successfully. Please check your email to verify your account.")
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user with email and password, returns JWT token. Email must be verified.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      LoginRequest  true  "Login credentials"
// @Success      200         {object}  user.TokenResponseWrapper  "Login successful"
// @Failure      400         {object}  response.ErrorResponse  "Validation error"
// @Failure      401         {object}  response.ErrorResponse  "Invalid credentials"
// @Failure      403         {object}  response.ErrorResponse  "Email not verified"
// @Router       /auth/login [post]
func (h *Handler) LoginHandler(c *gin.Context) {
	h.Login(c, h.jwtSecret, h.tokenExpiry)
}

// VerifyEmailHandler godoc
// @Summary      Verify email address
// @Description  Verify user email address using the verification token sent via email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        token  query     string  true  "Email verification token"
// @Success      200    {object}  response.Response  "Email verified successfully"
// @Failure      400    {object}  response.ErrorResponse  "Token is required"
// @Failure      404    {object}  response.ErrorResponse  "Invalid or expired verification token"
// @Failure      500    {object}  response.ErrorResponse  "Internal server error"
// @Router       /auth/verify-email [get]
func (h *Handler) VerifyEmailHandler(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.BadRequest(c, "Verification token is required", nil)
		return
	}

	// Find user by verification token
	userToVerify, exists := h.repo.GetByVerificationToken(token)
	if !exists {
		response.NotFound(c, "Invalid or expired verification token")
		return
	}

	// Update user status to active and clear verification token
	userToVerify.Status = StatusActive
	userToVerify.EmailVerificationToken = ""

	// Update user in repository
	updated, err := h.repo.Update(c.Request.Context(), userToVerify.ID, userToVerify)
	if err != nil {
		response.InternalError(c, "Failed to verify email: "+err.Error())
		return
	}

	// Remove sensitive fields from response
	updated.Password = ""
	updated.EmailVerificationToken = ""

	response.Success(c, http.StatusOK, updated, "Email verified successfully. You can now log in.")
}

// ListUsers godoc
// @Summary      List all users
// @Description  Get a list of all users (requires admin role)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  user.UsersResponseWrapper  "Users retrieved successfully"
// @Failure      401  {object}  response.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  response.ErrorResponse  "Forbidden"
// @Router       /users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	users := h.repo.List()
	// Remove passwords from response
	for i := range users {
		users[i].Password = ""
	}
	response.Success(c, http.StatusOK, users, "Users retrieved successfully")
}

// CreateUser godoc
// @Summary      Create a new user
// @Description  Create a new user with email, password, and optional role (requires admin role). A verification email will be sent to the provided email address.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user  body      createUserPayload  true  "User payload"
// @Success      201   {object}  user.UserResponseWrapper  "User created successfully. Verification email sent."
// @Failure      400   {object}  response.ErrorResponse  "Validation error"
// @Failure      401   {object}  response.ErrorResponse  "Unauthorized"
// @Failure      403   {object}  response.ErrorResponse  "Forbidden"
// @Failure      409   {object}  response.ErrorResponse  "Email already registered"
// @Failure      500   {object}  response.ErrorResponse  "Internal server error"
// @Router       /users [post]
func (h *Handler) CreateUser(c *gin.Context) {
	var payload createUserPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ValidationError(c, "Invalid request data", map[string]interface{}{
			"validation_error": err.Error(),
		})
		return
	}

	// Hash the password
	hashedPassword, err := HashPassword(payload.Password)
	if err != nil {
		response.InternalError(c, "Failed to hash password")
		return
	}

	// Validate and set role
	role := payload.Role
	if role == "" {
		role = RoleUser // Default role
	} else {
		// Validate role
		validRole := false
		for _, valid := range ValidRoles {
			if role == valid {
				validRole = true
				break
			}
		}
		if !validRole {
			response.ValidationError(c, "Invalid role", map[string]interface{}{
				"valid_roles": ValidRoles,
			})
			return
		}
	}

	// Generate verification token
	verificationToken, err := email.GenerateVerificationToken()
	if err != nil {
		response.InternalError(c, "Failed to generate verification token")
		return
	}

	// Create user with pending status (requires email verification)
	user := User{
		Email:                  payload.Email,
		Password:               hashedPassword,
		Role:                   role,
		Status:                 StatusPending,
		EmailVerificationToken: verificationToken,
	}

	// Create in repository (repository handles Airtable sync if configured)
	created, err := h.repo.Create(c.Request.Context(), user)
	if err != nil {
		// Check if it's a duplicate email error
		if strings.Contains(err.Error(), "already exists") {
			response.DuplicateEmail(c)
			return
		}
		response.InternalError(c, "Failed to create user: "+err.Error())
		return
	}

	// Send verification email
	if h.emailService != nil {
		if err := h.emailService.SendVerificationEmail(created.Email, verificationToken, h.baseURL); err != nil {
			// Log error but don't fail user creation - email can be resent later
			c.Error(err) // Log to gin context for error tracking
		}
	}

	// Remove sensitive fields from response
	created.Password = ""
	created.EmailVerificationToken = ""

	response.Success(c, http.StatusCreated, created, "User created successfully. Verification email sent to the user.")
}

// DeleteUser godoc
// @Summary      Delete a user by ID
// @Description  Delete a user using its ID (requires admin role)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "User ID"
// @Success      200  {object}  response.Response  "User deleted successfully"
// @Failure      401  {object}  response.ErrorResponse  "Unauthorized"
// @Failure      403  {object}  response.ErrorResponse  "Forbidden"
// @Failure      404  {object}  response.ErrorResponse  "User not found"
// @Router       /users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "User ID is required", nil)
		return
	}

	if ok := h.repo.Delete(id); !ok {
		response.NotFound(c, "User")
		return
	}

	response.SuccessNoContent(c, "User deleted successfully")
}

// UpdateUser godoc
// @Summary      Update user role and password
// @Description  Update a user's role and/or password by ID (requires super admin role)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string              true  "User ID"
// @Param        user  body      updateUserPayload  true  "Update payload (role and/or password)"
// @Success      200   {object}  user.UserResponseWrapper  "User updated successfully"
// @Failure      400   {object}  response.ErrorResponse  "Validation error"
// @Failure      401   {object}  response.ErrorResponse  "Unauthorized"
// @Failure      403   {object}  response.ErrorResponse  "Forbidden"
// @Failure      404   {object}  response.ErrorResponse  "User not found"
// @Failure      500   {object}  response.ErrorResponse  "Internal server error"
// @Router       /users/{id} [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "User ID is required", nil)
		return
	}

	var payload updateUserPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ValidationError(c, "Invalid request data", map[string]interface{}{
			"validation_error": err.Error(),
		})
		return
	}

	// Get existing user
	existingUser, exists := h.repo.Get(id)
	if !exists {
		response.NotFound(c, "User")
		return
	}

	// Prepare update
	updatedUser := existingUser

	// Update password if provided
	if payload.Password != "" {
		if len(payload.Password) < 6 {
			response.ValidationError(c, "Password must be at least 6 characters", nil)
			return
		}
		hashedPassword, err := HashPassword(payload.Password)
		if err != nil {
			response.InternalError(c, "Failed to hash password")
			return
		}
		updatedUser.Password = hashedPassword
	}

	// Update role if provided
	if payload.Role != "" {
		// Validate role
		validRole := false
		for _, valid := range ValidRoles {
			if payload.Role == valid {
				validRole = true
				break
			}
		}
		if !validRole {
			response.ValidationError(c, "Invalid role", map[string]interface{}{
				"valid_roles": ValidRoles,
			})
			return
		}
		updatedUser.Role = payload.Role
	}

	// Check if at least one field is being updated
	if payload.Password == "" && payload.Role == "" {
		response.ValidationError(c, "At least one field (password or role) must be provided", nil)
		return
	}

	// Update in repository (repository handles Airtable sync if configured)
	updated, err := h.repo.Update(c.Request.Context(), id, updatedUser)
	if err != nil {
		response.InternalError(c, "Failed to update user: "+err.Error())
		return
	}

	// Remove password from response
	updated.Password = ""
	response.Success(c, http.StatusOK, updated, "User updated successfully")
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type createUserPayload struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"`
}

type updateUserPayload struct {
	Password string `json:"password"` // Optional, min 6 characters if provided
	Role     string `json:"role"`     // Optional, must be valid role if provided
}
