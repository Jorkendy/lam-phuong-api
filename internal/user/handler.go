package user

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler exposes HTTP handlers for the user resource
type Handler struct {
	repo        Repository
	jwtSecret   string
	tokenExpiry time.Duration
}

// NewHandler creates a handler with the provided repository
func NewHandler(repo Repository, jwtSecret string, tokenExpiry time.Duration) *Handler {
	return &Handler{
		repo:        repo,
		jwtSecret:   jwtSecret,
		tokenExpiry: tokenExpiry,
	}
}

// RegisterRoutes attaches user routes to the supplied router group
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// Public routes
	router.POST("/auth/login", h.LoginHandler)
	router.POST("/users", h.CreateUser)

	// Protected routes
	protected := router.Group("")
	protected.Use(AuthMiddleware(h.jwtSecret))
	{
		protected.GET("/users", h.ListUsers)
		// protected.POST("/users", h.CreateUser)
		protected.DELETE("/users/:id", h.DeleteUser)
	}
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user with email and password, returns JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        credentials  body      LoginRequest  true  "Login credentials"
// @Success      200         {object}  TokenResponse
// @Failure      400         {object}  map[string]string
// @Failure      401         {object}  map[string]string
// @Router       /auth/login [post]
func (h *Handler) LoginHandler(c *gin.Context) {
	h.Login(c, h.jwtSecret, h.tokenExpiry)
}

// ListUsers godoc
// @Summary      List all users
// @Description  Get a list of all users (requires authentication)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   User
// @Failure      401  {object}  map[string]string
// @Router       /users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	users := h.repo.List()
	// Remove passwords from response
	for i := range users {
		users[i].Password = ""
	}
	c.JSON(http.StatusOK, users)
}

// CreateUser godoc
// @Summary      Create a new user
// @Description  Create a new user with email and password (requires authentication)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        user  body      createUserPayload  true  "User payload"
// @Success      201   {object}  User
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /users [post]
func (h *Handler) CreateUser(c *gin.Context) {
	var payload createUserPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash the password
	hashedPassword, err := HashPassword(payload.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := User{
		Email:    payload.Email,
		Password: hashedPassword,
	}

	// Create in repository (repository handles Airtable sync if configured)
	created, err := h.repo.Create(c.Request.Context(), user)
	if err != nil {
		// Check if it's a duplicate email error
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Remove password from response
	created.Password = ""
	c.JSON(http.StatusCreated, created)
}

// DeleteUser godoc
// @Summary      Delete a user by ID
// @Description  Delete a user using its ID (requires authentication)
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path      string  true  "User ID"
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if ok := h.repo.Delete(id); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

type createUserPayload struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}
