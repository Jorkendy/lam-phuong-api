package location

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler exposes HTTP handlers for the location resource.
type Handler struct {
	repo Repository
}

// NewHandler creates a handler with the provided repository.
func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes attaches location routes to the supplied router group.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/locations", h.ListLocations)
	router.GET("/locations/:id", h.GetLocation)
	router.POST("/locations", h.CreateLocation)
	router.PUT("/locations/:id", h.UpdateLocation)
	router.DELETE("/locations/:id", h.DeleteLocation)
}

func (h *Handler) ListLocations(c *gin.Context) {
	c.JSON(http.StatusOK, h.repo.List())
}

func (h *Handler) GetLocation(c *gin.Context) {
	id := c.Param("id")
	location, ok := h.repo.Get(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "location not found"})
		return
	}

	c.JSON(http.StatusOK, location)
}

func (h *Handler) CreateLocation(c *gin.Context) {
	var payload locationPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	location := Location{
		Name:    payload.Name,
		Address: payload.Address,
		City:    payload.City,
	}

	created, err := h.repo.Create(location)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *Handler) UpdateLocation(c *gin.Context) {
	id := c.Param("id")

	var payload locationPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	location := Location{
		Name:    payload.Name,
		Address: payload.Address,
		City:    payload.City,
	}

	updated, ok := h.repo.Update(id, location)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "location not found"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *Handler) DeleteLocation(c *gin.Context) {
	id := c.Param("id")
	if ok := h.repo.Delete(id); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "location not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

type locationPayload struct {
	Name    string `json:"name" binding:"required"`
	Address string `json:"address" binding:"required"`
	City    string `json:"city" binding:"required"`
}
