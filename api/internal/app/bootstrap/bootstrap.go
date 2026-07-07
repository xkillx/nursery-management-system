package bootstrap

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nursery-management-system/api/internal/platform/audit"
	httpserver "nursery-management-system/api/internal/platform/http"
	"nursery-management-system/api/internal/platform/version"

	paymentsapp "nursery-management-system/api/internal/modules/payments/application"
	paymentsdomain "nursery-management-system/api/internal/modules/payments/domain"
	"nursery-management-system/api/internal/platform/transaction"
)

type healthPinger interface {
	Ping(context.Context) error
}

type txManagerAdapter struct {
	mgr *transaction.Manager
}

func (a *txManagerAdapter) ExecTx(ctx context.Context, fn func(tx paymentsdomain.Tx) error) error {
	return a.mgr.ExecTx(ctx, func(tx pgx.Tx) error { return fn(tx) })
}

func registerHealthRoutes(router *gin.Engine, basePath string, pinger healthPinger) *gin.RouterGroup {
	router.GET("/health", healthHandler(pinger))
	router.GET("/version", version.Handler())
	api := router.Group(basePath)
	api.GET("/health", healthHandler(pinger))
	api.GET("/version", version.Handler())
	return api
}

func healthHandler(pinger healthPinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := pinger.Ping(ctx); err != nil {
			httpserver.WriteError(c, http.StatusServiceUnavailable, "db_unavailable", "Database is unavailable.", nil)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":     "ok",
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"request_id": httpserver.RequestIDFromContext(c),
		})
	}
}

type auditSystemWriterAdapter struct {
	w *audit.Writer
}

func (a *auditSystemWriterAdapter) WriteSystemWithTx(ctx context.Context, tx paymentsdomain.Tx, tenantID, branchID uuid.UUID, requestID string, params paymentsapp.SystemAuditParams) error {
	var reasonCode *string
	if params.ReasonCode != nil {
		reasonCode = params.ReasonCode
	}
	return a.w.WriteSystemWithTx(ctx, tx.(pgx.Tx), tenantID, branchID, requestID, audit.WriteParams{
		ActionType: params.ActionType,
		EntityType: params.EntityType,
		EntityID:   params.EntityID,
		ReasonCode: reasonCode,
		Details:    params.Details,
	})
}
