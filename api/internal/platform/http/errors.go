package httpserver

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	Path      string      `json:"path"`
	RequestID string      `json:"request_id"`
	Timestamp string      `json:"timestamp"`
}

func WriteError(c *gin.Context, status int, code, message string, details interface{}) {
	resp := ErrorResponse{
		Code:      code,
		Message:   message,
		Details:   details,
		Path:      c.Request.URL.Path,
		RequestID: requestIDFromContext(c),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	c.AbortWithStatusJSON(status, resp)
}

func writeError(c *gin.Context, status int, code, message string, details interface{}) {
	WriteError(c, status, code, message, details)
}

func writeInternalError(c *gin.Context) {
	WriteError(c, http.StatusInternalServerError, "internal_error", "Something went wrong.", nil)
}

func WriteInternalError(c *gin.Context) {
	writeInternalError(c)
}
