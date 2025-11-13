package location

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
)

// Handler exposes HTTP handlers for the location resource.
type Handler struct {
	repo Repository
}

// NewHandler creates a handler with the provided repository.
func NewHandler(repo Repository) *Handler {
	return &Handler{
		repo: repo,
	}
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

	// Generate slug from name if not provided
	locationSlug := payload.Slug
	if locationSlug != "" {
		locationSlug = slug.Make(locationSlug)
	} else {
		locationSlug = slug.Make(payload.Name)
	}

	locationSlug = ensureUniqueSlug(h.repo, locationSlug)

	// Validate status if provided, default to active
	status := StatusActive
	if payload.Status != "" {
		status = Status(payload.Status)
		if !status.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status. must be 'Active' or 'Disabled'"})
			return
		}
	}

	location := Location{
		Name:   payload.Name,
		Slug:   locationSlug,
		Status: status,
	}

	// Create in repository (repository handles Airtable sync if configured)
	created, err := h.repo.Create(c.Request.Context(), location)
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

	// Validate status if provided
	var status Status
	if payload.Status != "" {
		status = Status(payload.Status)
		if !status.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status. must be 'active' or 'disabled'"})
			return
		}
	} else {
		// Get existing location to preserve status if not provided
		existing, ok := h.repo.Get(id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "location not found"})
			return
		}
		status = existing.Status
	}

	location := Location{
		Name:   payload.Name,
		Slug:   payload.Slug,
		Status: status,
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
	Name   string `json:"name" binding:"required"` // Required
	Slug   string `json:"slug"`                    // Optional, will be generated from name if not provided
	Status string `json:"status"`                  // Optional, defaults to "Active"
}

func ensureUniqueSlug(repo Repository, baseSlug string) string {
	if baseSlug == "" {
		baseSlug = "location"
	}

	existingSlugs := make(map[string]struct{})
	for _, loc := range repo.List() {
		existingSlugs[loc.Slug] = struct{}{}
	}

	if _, exists := existingSlugs[baseSlug]; !exists {
		return baseSlug
	}

	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s-%d", baseSlug, i)
		if _, exists := existingSlugs[candidate]; !exists {
			return candidate
		}
	}
}
