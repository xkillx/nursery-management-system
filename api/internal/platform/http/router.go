package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"nursery-management-system/api/internal/platform/config"
)

type healthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id"`
}

func NewRouter(cfg config.Config, logger *slog.Logger, pool *pgxpool.Pool) *gin.Engine {
	if cfg.AppEnv == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(requestIDMiddleware())
	router.Use(accessLogMiddleware(logger))
	router.Use(recoveryMiddleware(logger))

	router.GET("/health", healthHandler(pool))
	api := router.Group(cfg.APIBasePath)
	api.GET("/health", healthHandler(pool))

	return router
}

func healthHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			writeError(c, http.StatusServiceUnavailable, "db_unavailable", "Database is unavailable.", nil)
			return
		}

		c.JSON(http.StatusOK, healthResponse{
			Status:    "ok",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			RequestID: requestIDFromContext(c),
		})
	}
}
