package server

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"lam-phuong-api/internal/email"
	"lam-phuong-api/internal/location"
	productGroup "lam-phuong-api/internal/productGroup"
	"lam-phuong-api/internal/response"
	"lam-phuong-api/internal/user"
)

// NewRouter constructs a Gin engine configured with middleware and routes.
func NewRouter(locationHandler *location.Handler, productGroupHandler *productGroup.Handler, userHandler *user.Handler, emailHandler *email.Handler, jwtSecret string) *gin.Engine {
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

	// ✅ THÊM HEALTH ENDPOINT
	router.GET("/health", func(c *gin.Context) {
		response.Success(c, 200, gin.H{
			"status":  "ok",
			"version": "1.0.0-alpha.0",
		}, "Service is healthy")
	})

	api := router.Group("/api")
	{
		// Auth routes (public)
		userHandler.RegisterRoutes(api)

		// Email test route (public)
		api.POST("/email/test", func(c *gin.Context) {
			if emailHandler == nil {
				response.InternalError(c, "Email service is not configured. Please set EMAIL_CLIENT_ID, EMAIL_CLIENT_SECRET, and EMAIL_REFRESH_TOKEN environment variables.")
				return
			}
			emailHandler.SendTestEmail(c)
		})

		// Protected routes (require authentication)
		protected := api.Group("")
		protected.Use(user.AuthMiddleware(jwtSecret))
		{
			// User password change route (authenticated users - own password only)
			protected.POST("/auth/change-password", userHandler.ChangePassword)

			// User password change by ID (Super Admin can change any, others can only change own)
			protected.POST("/users/:id/change-password", userHandler.ChangeUserPassword)

			// User management routes (admin only)
			adminRoutes := protected.Group("")
			adminRoutes.Use(user.RequireAdmin())
			{
				adminRoutes.GET("/users", userHandler.ListUsers)
				adminRoutes.POST("/users", userHandler.CreateUser)
				adminRoutes.DELETE("/users/:id", userHandler.DeleteUser)
				adminRoutes.POST("/users/:id/toggle-status", userHandler.ToggleUserStatus)
			}

			// User update routes (super admin only)
			superAdminRoutes := protected.Group("")
			superAdminRoutes.Use(user.RequireRole(user.RoleSuperAdmin))
			{
				superAdminRoutes.PUT("/users/:id", userHandler.UpdateUser)
			}

			// Location routes (authenticated users)
			locationHandler.RegisterRoutes(protected)

			// Product group routes (authenticated users)
			productGroupHandler.RegisterRoutes(protected)
		}
	}

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
