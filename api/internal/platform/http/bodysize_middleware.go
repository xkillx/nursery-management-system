package httpserver

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// BodySizeLimitMiddleware wraps request bodies with http.MaxBytesReader so that
// reading beyond maxBytes triggers an error. Skip paths bypass the wrapper
// (useful for endpoints that manage their own limits, e.g. Stripe webhooks).
func BodySizeLimitMiddleware(maxBytes int64, skipPaths map[string]bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if skipPaths[c.FullPath()] {
			c.Next()
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)

		c.Next()

		for _, ginErr := range c.Errors {
			var maxErr *http.MaxBytesError
			if errors.As(ginErr.Err, &maxErr) {
				WriteError(c, http.StatusRequestEntityTooLarge, "payload_too_large", "Request body too large.", nil)
				return
			}
		}
	}
}
