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

func writeError(c *gin.Context, status int, code, message string, details interface{}) {
	resp := ErrorResponse{
		Code:      code,
		Message:   message,
		Details:   details,
		RequestID: requestIDFromContext(c),
	}

	c.AbortWithStatusJSON(status, resp)
}

func writeInternalError(c *gin.Context) {
	writeError(c, http.StatusInternalServerError, "internal_error", "Something went wrong.", nil)
}
