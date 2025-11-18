package email

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"lam-phuong-api/internal/response"
)

// Handler handles email-related routes
type Handler struct {
	service *Service
}

// NewHandler creates a new email handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// TestEmailRequest represents the test email request payload
type TestEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// SendTestEmail godoc
// @Summary      Send test email
// @Description  Send a test email to the specified email address
// @Tags         email
// @Accept       json
// @Produce      json
// @Param        request  body      email.TestEmailRequest  true  "Test email request"
// @Success      200      {object}  response.Response  "Test email sent successfully"
// @Failure      400      {object}  response.ErrorResponse  "Validation error"
// @Failure      500      {object}  response.ErrorResponse  "Internal server error"
// @Router       /email/test [post]
func (h *Handler) SendTestEmail(c *gin.Context) {
	var req TestEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request data", map[string]interface{}{
			"validation_error": err.Error(),
		})
		return
	}

	// Send test email
	if err := h.service.SendTestEmail(req.Email); err != nil {
		response.InternalError(c, "Failed to send test email: "+err.Error())
		return
	}

	response.Success(c, http.StatusOK, gin.H{
		"email": req.Email,
	}, "Test email sent successfully")
}

