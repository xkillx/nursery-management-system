package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	RequestID string      `json:"request_id"`
}

func WriteError(c *gin.Context, status int, code, message string, details interface{}) {
	resp := ErrorResponse{
		Code:      code,
		Message:   message,
		Details:   details,
		RequestID: requestIDFromContext(c),
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
