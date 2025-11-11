package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"lam-phuong-api/internal/book"
	"lam-phuong-api/internal/location"
)

// NewRouter constructs a Gin engine configured with middleware and routes.
func NewRouter(bookHandler *book.Handler, locationHandler *location.Handler) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api")
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		bookHandler.RegisterRoutes(api)
		locationHandler.RegisterRoutes(api)
	}

	return router
}
