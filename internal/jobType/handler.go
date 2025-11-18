package jobtype

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"lam-phuong-api/internal/response"
	"lam-phuong-api/internal/user"
)

// Handler exposes HTTP handlers for the job type resource.
type Handler struct {
	repo Repository
}

// NewHandler creates a handler with the provided repository.
func NewHandler(repo Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

// RegisterRoutes attaches job type routes to the supplied router group.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/job-types", h.ListJobTypes)
	router.POST("/job-types", h.CreateJobType)
	router.DELETE("/job-types/:slug", h.DeleteJobTypeBySlug)
	router.POST("/job-types/:slug/toggle-status", h.ToggleJobTypeStatus)
}

// ListJobTypes godoc
// @Summary      List all job types
// @Description  Get a list of all job types (requires authentication)
// @Tags         job-types
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  jobtype.JobTypesResponseWrapper  "Job types retrieved successfully"
// @Failure      401  {object}  response.ErrorResponse  "Unauthorized"
// @Router       /job-types [get]
func (h *Handler) ListJobTypes(c *gin.Context) {
	jobTypes := h.repo.List()
	response.Success(c, http.StatusOK, jobTypes, "Job types retrieved successfully")
}

// CreateJobType godoc
// @Summary      Create a new job type
// @Description  Create a new job type with name and optional slug. If slug is not provided, it will be generated from the name. (requires authentication)
// @Tags         job-types
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        jobType  body      jobTypePayload  true  "Job type payload"
// @Success      201      {object}  jobtype.JobTypeResponseWrapper  "Job type created successfully"
// @Failure      400      {object}  response.ErrorResponse  "Validation error"
// @Failure      401      {object}  response.ErrorResponse  "Unauthorized"
// @Failure      500      {object}  response.ErrorResponse  "Internal server error"
// @Router       /job-types [post]
func (h *Handler) CreateJobType(c *gin.Context) {
	var payload jobTypePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ValidationError(c, "Invalid request data", map[string]interface{}{
			"validation_error": err.Error(),
		})
		return
	}

	// Generate slug from name if not provided
	jobTypeSlug := payload.Slug
	if jobTypeSlug != "" {
		jobTypeSlug = slug.Make(jobTypeSlug)
	} else {
		jobTypeSlug = slug.Make(payload.Name)
	}

	jobTypeSlug = ensureUniqueSlug(h.repo, jobTypeSlug)

	jobType := JobType{
		Name: payload.Name,
		Slug: jobTypeSlug,
	}

	// Create in repository (repository handles Airtable sync if configured)
	created, err := h.repo.Create(c.Request.Context(), jobType)
	if err != nil {
		response.InternalError(c, "Failed to create job type: "+err.Error())
		return
	}

	response.Success(c, http.StatusCreated, created, "Job type created successfully")
}

type jobTypePayload struct {
	Name string `json:"name" binding:"required"` // Required
	Slug string `json:"slug"`                     // Optional, will be generated from name if not provided
}

func ensureUniqueSlug(repo Repository, baseSlug string) string {
	if baseSlug == "" {
		baseSlug = "job-type"
	}

	existingSlugs := make(map[string]struct{})
	for _, jt := range repo.List() {
		existingSlugs[jt.Slug] = struct{}{}
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

// DeleteJobTypeBySlug godoc
// @Summary      Delete a job type by slug
// @Description  Delete a job type using its slug (requires authentication)
// @Tags         job-types
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        slug  path      string  true  "Job type slug"
// @Success      200   {object}  response.Response  "Job type deleted successfully"
// @Failure      400   {object}  response.ErrorResponse  "Validation error"
// @Failure      401   {object}  response.ErrorResponse  "Unauthorized"
// @Failure      404   {object}  response.ErrorResponse  "Job type not found"
// @Router       /job-types/{slug} [delete]
func (h *Handler) DeleteJobTypeBySlug(c *gin.Context) {
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
		response.NotFound(c, "Job type")
		return
	}

	response.SuccessNoContent(c, "Job type deleted successfully")
}

// ToggleJobTypeStatus godoc
// @Summary      Toggle job type status
// @Description  Toggle a job type's status between Active and Disabled. Only Admin or Super Admin can call this endpoint.
// @Tags         job-types
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        slug  path      string  true  "Job type slug"
// @Success      200   {object}  jobtype.JobTypeResponseWrapper  "Job type status toggled successfully"
// @Failure      400   {object}  response.ErrorResponse  "Validation error"
// @Failure      401   {object}  response.ErrorResponse  "Unauthorized"
// @Failure      403   {object}  response.ErrorResponse  "Forbidden - Admin or Super Admin role required"
// @Failure      404   {object}  response.ErrorResponse  "Job type not found"
// @Failure      500   {object}  response.ErrorResponse  "Internal server error"
// @Router       /job-types/{slug}/toggle-status [post]
func (h *Handler) ToggleJobTypeStatus(c *gin.Context) {
	// Check if user has admin role
	userRole, exists := c.Get("user_role")
	if !exists {
		response.Unauthorized(c, "User role not found")
		return
	}
	role := userRole.(string)
	if role != user.RoleAdmin && role != user.RoleSuperAdmin {
		response.Forbidden(c, "Admin or Super Admin role required")
		return
	}

	slugParam := c.Param("slug")
	if slugParam == "" {
		response.BadRequest(c, "Job type slug is required", nil)
		return
	}

	normalizedSlug := slug.Make(slugParam)
	if normalizedSlug == "" {
		response.ValidationError(c, "Invalid slug format", nil)
		return
	}

	// Get existing job type by slug
	existingJobType, exists := h.repo.GetBySlug(normalizedSlug)
	if !exists {
		response.NotFound(c, "Job type")
		return
	}

	// Toggle status between Active and Disabled
	var newStatus string
	if existingJobType.Status == StatusActive {
		newStatus = StatusDisabled
	} else {
		newStatus = StatusActive
	}

	// Update job type status
	existingJobType.Status = newStatus
	updated, err := h.repo.Update(c.Request.Context(), existingJobType.ID, existingJobType)
	if err != nil {
		response.InternalError(c, "Failed to update job type status: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, updated, "Job type status toggled successfully")
}

