package pagination

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestContext(query string) *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/?"+query, nil)
	return c
}

func TestParsePageParams_HappyPath(t *testing.T) {
	c := setupTestContext("page=2&page_size=25")
	page, pageSize := ParsePageParams(c)
	assert.Equal(t, 2, page)
	assert.Equal(t, 25, pageSize)
}

func TestParsePageParams_Defaults(t *testing.T) {
	c := setupTestContext("")
	page, pageSize := ParsePageParams(c)
	assert.Equal(t, 1, page)
	assert.Equal(t, 50, pageSize)
}

func TestParsePageParams_BoundaryPageSizeZero(t *testing.T) {
	c := setupTestContext("page_size=0")
	_, pageSize := ParsePageParams(c)
	assert.Equal(t, 50, pageSize) // 0 returns default
}

func TestParsePageParams_BoundaryPageSizeOverMax(t *testing.T) {
	c := setupTestContext("page_size=300")
	_, pageSize := ParsePageParams(c)
	assert.Equal(t, 200, pageSize)
}

func TestParsePageParams_BoundaryPageZero(t *testing.T) {
	c := setupTestContext("page=0")
	page, _ := ParsePageParams(c)
	assert.Equal(t, 1, page)
}

func TestParsePageParams_InvalidInput(t *testing.T) {
	c := setupTestContext("page=abc&page_size=xyz")
	page, pageSize := ParsePageParams(c)
	assert.Equal(t, 1, page)
	assert.Equal(t, 50, pageSize)
}

func TestPaginatedResponse(t *testing.T) {
	items := []string{"a", "b", "c"}
	result := PaginatedResponse(items, 120, 2, 25)

	assert.Equal(t, items, result["items"])
	assert.Equal(t, 120, result["total"])
	assert.Equal(t, 2, result["page"])
	assert.Equal(t, 25, result["page_size"])
	assert.Equal(t, 5, result["pages"])
}

func TestPaginatedResponse_PagesZeroTotal(t *testing.T) {
	result := PaginatedResponse([]string{}, 0, 1, 50)
	assert.Equal(t, 0, result["pages"])
}

func TestPaginatedResponse_PagesSingleItem(t *testing.T) {
	result := PaginatedResponse([]string{"a"}, 1, 1, 50)
	assert.Equal(t, 1, result["pages"])
}

func TestPaginatedResponse_PagesExactMultiple(t *testing.T) {
	result := PaginatedResponse([]string{}, 50, 1, 50)
	assert.Equal(t, 1, result["pages"])
}

func TestPaginatedResponse_PagesCeilingDivision(t *testing.T) {
	result := PaginatedResponse([]string{}, 51, 1, 50)
	assert.Equal(t, 2, result["pages"])
}
