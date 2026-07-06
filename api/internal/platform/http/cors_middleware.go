package httpserver

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware returns a gin.HandlerFunc that handles CORS with configurable allowed origins.
// allowedOrigins is a list of specific origins (e.g., ["https://app.example.com"]).
// When empty, falls back to webBaseURL.
func CORSMiddleware(allowedOrigins []string, webBaseURL string) gin.HandlerFunc {
	origins := allowedOrigins
	if len(origins) == 0 && webBaseURL != "" {
		origins = []string{webBaseURL}
	}

	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// ParseAllowedOrigins splits a comma-separated origins string into a slice.
func ParseAllowedOrigins(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s != "" {
			origins = append(origins, s)
		}
	}
	return origins
}
