package httpserver

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// errorChain walks the Unwrap chain and returns each level's message.
// The first element is the outermost error; the last is the root cause.
func errorChain(err error) []string {
	var chain []string
	for e := err; e != nil; e = errors.Unwrap(e) {
		chain = append(chain, e.Error())
		// Guard against self-unwrapping sentinels to avoid infinite loops.
		if u := errors.Unwrap(e); u == e {
			break
		}
	}
	return chain
}

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
		args = append(args, "error_chain", errorChain(err))
	}

	logger.Error("request_failed", args...)
}
