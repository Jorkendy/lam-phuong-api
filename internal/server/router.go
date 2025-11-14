package server

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"lam-phuong-api/internal/location"
	"lam-phuong-api/internal/user"
)

// NewRouter constructs a Gin engine configured with middleware and routes.
func NewRouter(locationHandler *location.Handler, userHandler *user.Handler, jwtSecret string) *gin.Engine {
	router := gin.Default()

	// Configure CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Allow all origins for development
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false, // Set to false when using wildcard origins
		MaxAge:           12 * time.Hour,
	}))

	api := router.Group("/api")
	{
		// Public routes
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		// Auth routes (public)
		userHandler.RegisterRoutes(api)

		// Protected routes (require authentication)
		protected := api.Group("")
		protected.Use(user.AuthMiddleware(jwtSecret))
		{
			// User management routes (admin only)
			adminRoutes := protected.Group("")
			adminRoutes.Use(user.RequireAdmin())
			{
				adminRoutes.GET("/users", userHandler.ListUsers)
				adminRoutes.POST("/users", userHandler.CreateUser)
				adminRoutes.DELETE("/users/:id", userHandler.DeleteUser)
			}

			// User update routes (super admin only)
			superAdminRoutes := protected.Group("")
			superAdminRoutes.Use(user.RequireRole(user.RoleSuperAdmin))
			{
				superAdminRoutes.PUT("/users/:id", userHandler.UpdateUser)
			}

			// Location routes (authenticated users)
			locationHandler.RegisterRoutes(protected)
		}
	}

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
