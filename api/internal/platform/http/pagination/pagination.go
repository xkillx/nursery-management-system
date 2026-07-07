package pagination

import (
	"github.com/gin-gonic/gin"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 50
	MaxPerPage     = 200
	MinPerPage     = 1
)

// ParsePageParams reads page and page_size query parameters from the request.
// Defaults: page=1, page_size=50. Clamps page_size to [1, 200] and page to >= 1.
func ParsePageParams(c *gin.Context) (page, pageSize int) {
	page = parseIntQuery(c, "page", DefaultPage)
	if page < 1 {
		page = 1
	}

	pageSize = parseIntQuery(c, "page_size", DefaultPerPage)
	if pageSize < MinPerPage {
		pageSize = MinPerPage
	}
	if pageSize > MaxPerPage {
		pageSize = MaxPerPage
	}

	return page, pageSize
}

// PaginatedResponse returns the standard pagination envelope.
func PaginatedResponse(items interface{}, total, page, pageSize int) gin.H {
	pages := 0
	if total > 0 && pageSize > 0 {
		pages = (total + pageSize - 1) / pageSize
	}
	return gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"pages":     pages,
	}
}

func parseIntQuery(c *gin.Context, key string, def int) int {
	v := c.Query(key)
	if v == "" {
		return def
	}
	var n int
	for _, r := range v {
		if r < '0' || r > '9' {
			return def
		}
		n = n*10 + int(r-'0')
	}
	if n == 0 {
		return def
	}
	return n
}
