package productgroup

import (
	"fmt"
	"net/http"

	"lam-phuong-api/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
)

// Handler exposes HTTP handlers for the product group resource.
type Handler struct {
	repo Repository
}

// NewHandler creates a handler with the provided repository.
func NewHandler(repo Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

// RegisterRoutes attaches product group routes to the supplied router group.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/product-groups", h.ListProductGroups)
	router.POST("/product-groups", h.CreateProductGroup)
	router.DELETE("/product-groups/:slug", h.DeleteProductGroupBySlug)
}

// ListProductGroups godoc
// @Summary      List all product groups
// @Description  Get a list of all product groups (requires authentication)
// @Tags         product-groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  productgroup.ProductGroupsResponseWrapper  "Product groups retrieved successfully"
// @Failure      401  {object}  response.ErrorResponse  "Unauthorized"
// @Router       /product-groups [get]
func (h *Handler) ListProductGroups(c *gin.Context) {
	productGroups := h.repo.List()
	response.Success(c, http.StatusOK, productGroups, "Product groups retrieved successfully")
}

// CreateProductGroup godoc
// @Summary      Create a new product group
// @Description  Create a new product group with name and optional slug. If slug is not provided, it will be generated from the name. (requires authentication)
// @Tags         product-groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        productGroup  body      productGroupPayload  true  "Product group payload"
// @Success      201           {object}  productgroup.ProductGroupResponseWrapper  "Product group created successfully"
// @Failure      400           {object}  response.ErrorResponse  "Validation error"
// @Failure      401           {object}  response.ErrorResponse  "Unauthorized"
// @Failure      500           {object}  response.ErrorResponse  "Internal server error"
// @Router       /product-groups [post]
func (h *Handler) CreateProductGroup(c *gin.Context) {
	var payload productGroupPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ValidationError(c, "Invalid request data", map[string]interface{}{
			"validation_error": err.Error(),
		})
		return
	}

	// Generate slug from name if not provided
	productGroupSlug := payload.Slug
	if productGroupSlug != "" {
		productGroupSlug = slug.Make(productGroupSlug)
	} else {
		productGroupSlug = slug.Make(payload.Name)
	}

	productGroupSlug = ensureUniqueSlug(h.repo, productGroupSlug)

	productGroup := ProductGroup{
		Name: payload.Name,
		Slug: productGroupSlug,
	}

	// Create in repository (repository handles Airtable sync if configured)
	created, err := h.repo.Create(c.Request.Context(), productGroup)
	if err != nil {
		response.InternalError(c, "Failed to create product group: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, created, "Product group created successfully")
}

type productGroupPayload struct {
	Name string `json:"name" binding:"required"` // Required
	Slug string `json:"slug"`                    // Optional, will be generated from name if not provided
}

func ensureUniqueSlug(repo Repository, baseSlug string) string {
	if baseSlug == "" {
		baseSlug = "product-group"
	}

	existingSlugs := make(map[string]struct{})
	for _, pg := range repo.List() {
		existingSlugs[pg.Slug] = struct{}{}
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

// DeleteProductGroupBySlug godoc
// @Summary      Delete a product group by slug
// @Description  Delete a product group using its slug (requires authentication)
// @Tags         product-groups
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        slug  path      string  true  "Product group slug"
// @Success      200   {object}  response.Response  "Product group deleted successfully"
// @Failure      400   {object}  response.ErrorResponse  "Validation error"
// @Failure      401   {object}  response.ErrorResponse  "Unauthorized"
// @Failure      404   {object}  response.ErrorResponse  "Product group not found"
// @Router       /product-groups/{slug} [delete]
func (h *Handler) DeleteProductGroupBySlug(c *gin.Context) {
	slugParam := c.Param("slug")
	if slugParam == "" {
		response.BadRequest(c, "Slug is required", nil)
		return
	}

	normalizedSlug := slug.Make(slugParam)
	if normalizedSlug == "" {
		response.ValidationError(c, "Invalid slug format", nil)
		return
	}

	if ok := h.repo.DeleteBySlug(normalizedSlug); !ok {
		response.NotFound(c, "Product group")
		return
	}

	response.SuccessNoContent(c, "Product group deleted successfully")
}
