package httpemail

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/email/application"
	"nursery-management-system/api/internal/modules/email/domain"
	"nursery-management-system/api/internal/platform/errors"
	"nursery-management-system/api/internal/platform/tenant"
)

type Handler struct {
	logger     *slog.Logger
	listEmails *application.ListEmails
	getEmail   *application.GetEmail
	retryEmail *application.RetryEmail
	getStats   *application.GetEmailStats
	repo       domain.OutboxRepository
}

func NewHandler(
	logger *slog.Logger,
	listEmails *application.ListEmails,
	getEmail *application.GetEmail,
	retryEmail *application.RetryEmail,
	getStats *application.GetEmailStats,
	repo domain.OutboxRepository,
) *Handler {
	return &Handler{
		logger:     logger,
		listEmails: listEmails,
		getEmail:   getEmail,
		retryEmail: retryEmail,
		getStats:   getStats,
		repo:       repo,
	}
}

func (h *Handler) RegisterRoutes(api *gin.RouterGroup) {
	email := api.Group("/email")
	email.GET("/outbox", h.ListEmails)
	email.GET("/outbox/:id", h.GetEmail)
	email.POST("/outbox/:id/retry", h.RetryEmail)
	email.GET("/stats", h.GetStats)
	email.POST("/webhooks/postmark", h.HandlePostmarkWebhook)
}

func (h *Handler) ListEmails(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials or session"})
		return
	}

	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	emails, total, err := h.listEmails.Execute(c.Request.Context(), actor.TenantID, actor.BranchID, statusPtr, limit, offset)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "list_emails_failed",
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list emails"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"emails": emails,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handler) GetEmail(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials or session"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email ID"})
		return
	}

	email, err := h.getEmail.Execute(c.Request.Context(), actor.TenantID, actor.BranchID, id)
	if err != nil {
		if errors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Email not found"})
			return
		}
		h.logger.ErrorContext(c.Request.Context(), "get_email_failed",
			"email_id", id,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get email"})
		return
	}

	c.JSON(http.StatusOK, email)
}

func (h *Handler) RetryEmail(c *gin.Context) {
	_, ok := tenant.ActorFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials or session"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email ID"})
		return
	}

	if err := h.retryEmail.Execute(c.Request.Context(), id); err != nil {
		if errors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Email not found"})
			return
		}
		h.logger.ErrorContext(c.Request.Context(), "retry_email_failed",
			"email_id", id,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retry email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email queued for retry"})
}

func (h *Handler) GetStats(c *gin.Context) {
	actor, ok := tenant.ActorFromGinContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials or session"})
		return
	}

	stats, err := h.getStats.Execute(c.Request.Context(), actor.TenantID, actor.BranchID)
	if err != nil {
		h.logger.ErrorContext(c.Request.Context(), "get_email_stats_failed",
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get email stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) HandlePostmarkWebhook(c *gin.Context) {
	h.logger.Warn("webhook_received_but_no_provider_configured",
		"message", "Webhook received but no provider configured for delivery tracking",
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook received",
	})
}
