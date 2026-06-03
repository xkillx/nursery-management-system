package httpserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/platform/metrics"
)

const (
	requestIDHeader     = "X-Request-ID"
	correlationIDHeader = "X-Correlation-ID"
	traceparentHeader   = "traceparent"

	requestIDContextKey     = "request_id"
	correlationIDContextKey = "correlation_id"
	traceIDContextKey       = "trace_id"

	maxHeaderLen = 128
)

var visibleASCII = regexp.MustCompile(`^[\x21-\x7E]+$`)

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := sanitizeRequestID(c.GetHeader(requestIDHeader))
		correlationID := sanitizeRequestID(c.GetHeader(correlationIDHeader))
		traceID := parseTraceparent(c.GetHeader(traceparentHeader))

		c.Set(requestIDContextKey, requestID)
		c.Set(correlationIDContextKey, correlationID)
		c.Set(traceIDContextKey, traceID)

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, requestIDContextKey, requestID)
		ctx = context.WithValue(ctx, correlationIDContextKey, correlationID)
		ctx = context.WithValue(ctx, traceIDContextKey, traceID)
		c.Request = c.Request.WithContext(ctx)

		c.Writer.Header().Set(requestIDHeader, requestID)
		c.Writer.Header().Set(correlationIDHeader, correlationID)
		c.Next()
	}
}

func accessLogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		logger.Info("http_request",
			"method", c.Request.Method,
			"route", route,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(startedAt).Milliseconds(),
			"client_ip", c.ClientIP(),
			"request_id", requestIDFromContext(c),
		)
	}
}

func AccessLogMiddlewareWithMetrics(logger *slog.Logger, recorder *metrics.Recorder) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		status := c.Writer.Status()
		latencyMs := time.Since(startedAt).Milliseconds()

		args := []any{
			"method", c.Request.Method,
			"route", route,
			"status", status,
			"status_class", statusClass(status),
			"latency_ms", latencyMs,
			"client_ip", c.ClientIP(),
			"request_id", requestIDFromContext(c),
			"correlation_id", CorrelationIDFromContext(c),
		}

		traceID := TraceIDFromContext(c)
		if traceID != "" {
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

		logger.Info("http_request", args...)

		recorder.ObserveHTTPRequest(route, c.Request.Method, status, float64(latencyMs)/1000.0)
	}
}

func statusClass(code int) string {
	switch {
	case code >= 100 && code < 200:
		return "1xx"
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500:
		return "5xx"
	default:
		return "unknown"
	}
}

func recoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger.Error("panic recovered", "recovered", recovered, "request_id", requestIDFromContext(c))
		writeInternalError(c)
	})
}

func RequestIDMiddleware() gin.HandlerFunc {
	return requestIDMiddleware()
}

func AccessLogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return accessLogMiddleware(logger)
}

func RecoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return recoveryMiddleware(logger)
}

func RequestIDFromContext(c *gin.Context) string {
	return requestIDFromContext(c)
}

func CorrelationIDFromContext(c *gin.Context) string {
	if v, ok := c.Get(correlationIDContextKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func TraceIDFromContext(c *gin.Context) string {
	if v, ok := c.Get(traceIDContextKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func RequestIDFromStdContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDContextKey).(string); ok {
		return v
	}
	return ""
}

func CorrelationIDFromStdContext(ctx context.Context) string {
	if v, ok := ctx.Value(correlationIDContextKey).(string); ok {
		return v
	}
	return ""
}

func TraceIDFromStdContext(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDContextKey).(string); ok {
		return v
	}
	return ""
}

func NewRequestID() string {
	return newRequestID()
}

func requestIDFromContext(c *gin.Context) string {
	if v, ok := c.Get(requestIDContextKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func newRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(buf)
}

func sanitizeRequestID(v string) string {
	v = trimSpaces(v)
	if len(v) == 0 || len(v) > maxHeaderLen || !visibleASCII.MatchString(v) {
		return newRequestID()
	}
	return v
}

func trimSpaces(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

// W3C traceparent format: version-traceid-spanid-flags
// version is 2 hex chars, traceid is 32 hex chars, spanid is 16 hex chars, flags is 2 hex chars
var traceparentRe = regexp.MustCompile(`^00-([0-9a-f]{32})-[0-9a-f]{16}-[0-9a-f]{2}$`)

func parseTraceparent(v string) string {
	v = trimSpaces(v)
	matches := traceparentRe.FindStringSubmatch(v)
	if len(matches) < 2 {
		return ""
	}
	traceID := matches[1]
	if traceID == "00000000000000000000000000000000" {
		return ""
	}
	return traceID
}
