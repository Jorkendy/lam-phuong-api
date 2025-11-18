package jobcategory

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"lam-phuong-api/internal/response"
)

// Handler exposes HTTP handlers for the job category resource.
type Handler struct {
	repo Repository
}

// NewHandler creates a handler with the provided repository.
func NewHandler(repo Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

// RegisterRoutes attaches job category routes to the supplied router group.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/job-categories", h.ListJobCategories)
	router.POST("/job-categories", h.CreateJobCategory)
	router.DELETE("/job-categories/:slug", h.DeleteJobCategoryBySlug)
}

// ListJobCategories godoc
// @Summary      List all job categories
// @Description  Get a list of all job categories (requires authentication)
// @Tags         job-categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  jobcategory.JobCategoriesResponseWrapper  "Job categories retrieved successfully"
// @Failure      401  {object}  response.ErrorResponse  "Unauthorized"
// @Router       /job-categories [get]
func (h *Handler) ListJobCategories(c *gin.Context) {
	jobCategories := h.repo.List()
	response.Success(c, http.StatusOK, jobCategories, "Job categories retrieved successfully")
}

// CreateJobCategory godoc
// @Summary      Create a new job category
// @Description  Create a new job category with name and optional slug. If slug is not provided, it will be generated from the name. (requires authentication)
// @Tags         job-categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        jobCategory  body      jobCategoryPayload  true  "Job category payload"
// @Success      201          {object}  jobcategory.JobCategoryResponseWrapper  "Job category created successfully"
// @Failure      400          {object}  response.ErrorResponse  "Validation error"
// @Failure      401          {object}  response.ErrorResponse  "Unauthorized"
// @Failure      500          {object}  response.ErrorResponse  "Internal server error"
// @Router       /job-categories [post]
func (h *Handler) CreateJobCategory(c *gin.Context) {
	var payload jobCategoryPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ValidationError(c, "Invalid request data", map[string]interface{}{
			"validation_error": err.Error(),
		})
		return
	}

	// Generate slug from name if not provided
	jobCategorySlug := payload.Slug
	if jobCategorySlug != "" {
		jobCategorySlug = slug.Make(jobCategorySlug)
	} else {
		jobCategorySlug = slug.Make(payload.Name)
	}

	jobCategorySlug = ensureUniqueSlug(h.repo, jobCategorySlug)

	jobCategory := JobCategory{
		Name: payload.Name,
		Slug: jobCategorySlug,
	}

	// Create in repository (repository handles Airtable sync if configured)
	created, err := h.repo.Create(c.Request.Context(), jobCategory)
	if err != nil {
		response.InternalError(c, "Failed to create job category: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, created, "Job category created successfully")
}

type jobCategoryPayload struct {
	Name string `json:"name" binding:"required"` // Required
	Slug string `json:"slug"`                     // Optional, will be generated from name if not provided
}

func ensureUniqueSlug(repo Repository, baseSlug string) string {
	if baseSlug == "" {
		baseSlug = "job-category"
	}

	existingSlugs := make(map[string]struct{})
	for _, jc := range repo.List() {
		existingSlugs[jc.Slug] = struct{}{}
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

// DeleteJobCategoryBySlug godoc
// @Summary      Delete a job category by slug
// @Description  Delete a job category using its slug (requires authentication)
// @Tags         job-categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        slug  path      string  true  "Job category slug"
// @Success      200   {object}  response.Response  "Job category deleted successfully"
// @Failure      400   {object}  response.ErrorResponse  "Validation error"
// @Failure      401   {object}  response.ErrorResponse  "Unauthorized"
// @Failure      404   {object}  response.ErrorResponse  "Job category not found"
// @Router       /job-categories/{slug} [delete]
func (h *Handler) DeleteJobCategoryBySlug(c *gin.Context) {
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
		response.NotFound(c, "Job category")
		return
	}

	response.SuccessNoContent(c, "Job category deleted successfully")
}

