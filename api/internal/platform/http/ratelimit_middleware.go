package httpserver

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"nursery-management-system/api/internal/platform/ratelimit"
)

const (
	rateLimitHeader  = "X-RateLimit-Limit"
	remainingHeader  = "X-RateLimit-Remaining"
	resetHeader      = "X-RateLimit-Reset"
	retryAfterHeader = "Retry-After"
)

func RateLimitMiddleware(limiter *ratelimit.FixedWindowLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		result := limiter.AllowWithInfo(ip)

		c.Header(rateLimitHeader, strconv.Itoa(result.Limit))
		c.Header(remainingHeader, strconv.Itoa(result.Remaining))
		c.Header(resetHeader, strconv.FormatInt(result.ResetAt.Unix(), 10))

		if !result.Allowed {
			retryAfter := int(time.Until(result.ResetAt).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header(retryAfterHeader, strconv.Itoa(retryAfter))
			WriteError(c, http.StatusTooManyRequests, "rate_limited", "Too many requests.", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
