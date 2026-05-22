package httpserver

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	requestIDHeader  = "X-Request-ID"
	requestIDContext = "request_id"
)

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = newRequestID()
		}

		c.Set(requestIDContext, requestID)
		c.Writer.Header().Set(requestIDHeader, requestID)
		c.Next()
	}
}

func accessLogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		logger.Info("http_request",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"latency_ms", time.Since(startedAt).Milliseconds(),
			"client_ip", c.ClientIP(),
			"request_id", requestIDFromContext(c),
		)
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

func requestIDFromContext(c *gin.Context) string {
	if v, ok := c.Get(requestIDContext); ok {
		if requestID, castOK := v.(string); castOK {
			return requestID
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
