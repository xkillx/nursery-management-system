package queryparams

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// SortExpression represents a parsed sort parameter.
type SortExpression struct {
	Field     string
	Direction string
}

// IsZero returns true when no sort parameter was provided (handler should use default ORDER BY).
func (s SortExpression) IsZero() bool {
	return s.Field == "" && s.Direction == ""
}

// ParseSortParams reads the "sort" query parameter in "field:direction" format
// and validates against the allowed map (field -> allowed directions).
// Returns a zero-value SortExpression when the parameter is omitted.
func ParseSortParams(c *gin.Context, allowed map[string][]string) (SortExpression, error) {
	raw := c.Query("sort")
	if raw == "" {
		return SortExpression{}, nil
	}

	parts := strings.SplitN(raw, ":", 2)
	if len(parts) != 2 {
		return SortExpression{}, fmt.Errorf("invalid sort format: expected field:direction, got %q", raw)
	}

	field := parts[0]
	direction := parts[1]

	allowedDirs, ok := allowed[field]
	if !ok {
		return SortExpression{}, fmt.Errorf("invalid sort field: %q is not sortable", field)
	}

	dirValid := false
	for _, d := range allowedDirs {
		if d == direction {
			dirValid = true
			break
		}
	}
	if !dirValid {
		return SortExpression{}, fmt.Errorf("invalid sort direction: %q is not allowed for field %q", direction, field)
	}

	return SortExpression{Field: field, Direction: direction}, nil
}
