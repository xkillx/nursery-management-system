package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// LogMappedError emits a single request_failed diagnostics entry for mapped
// internal errors. It is silent for status < 500 and when logger is nil so
// focused tests can opt out. It never logs request or response payloads,
// authorization headers, cookies, or SQL parameter values.
func LogMappedError(c *gin.Context, logger *slog.Logger, status int, code string, err error) {
	if logger == nil {
		return
	}
	if status < http.StatusInternalServerError {
		return
	}
	if c == nil {
		return
	}

	route := c.FullPath()
	if route == "" {
		route = "unmatched"
	}

	args := []any{
		"method", c.Request.Method,
		"route", route,
		"status", status,
		"status_class", statusClass(status),
		"error_code", code,
		"request_id", RequestIDFromContext(c),
		"correlation_id", CorrelationIDFromContext(c),
	}

	if traceID := TraceIDFromContext(c); traceID != "" {
		args = append(args, "trace_id", traceID)
	}

	if authCtx, ok := AuthContextFromContext(c); ok {
		args = append(args,
			"user_id", authCtx.UserID,
			"membership_id", authCtx.MembershipID,
			"tenant_id", authCtx.TenantID,
			"branch_id", authCtx.BranchID,
			"role", authCtx.Role,
		)
	}

	if err != nil {
		args = append(args, "error", err.Error())
	}

	logger.Error("request_failed", args...)
}
