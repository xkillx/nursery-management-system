package queryparams

import "github.com/gin-gonic/gin"

// ParseFilterParams extracts query parameters that match the allowed filter keys.
// allowed maps query parameter names to their expected type (e.g., "status" -> "string", "room_id" -> "uuid").
// Returns a map of matched parameter names to their string values.
func ParseFilterParams(c *gin.Context, allowed map[string]string) map[string]string {
	filters := make(map[string]string, len(allowed))
	for key := range allowed {
		val := c.Query(key)
		if val != "" {
			filters[key] = val
		}
	}
	return filters
}
