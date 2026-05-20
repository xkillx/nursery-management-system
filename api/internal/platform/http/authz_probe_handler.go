package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func meHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx, ok := authContextFromContext(c)
		if !ok {
			writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
			return
		}

		c.JSON(http.StatusOK, gin.H{"auth": authCtx})
	}
}

func scopeProbeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx, ok := authContextFromContext(c)
		if !ok {
			writeError(c, http.StatusUnauthorized, "unauthorized", "Invalid credentials or session.", nil)
			return
		}

		if c.Param("tenant_id") != authCtx.TenantID || c.Param("branch_id") != authCtx.BranchID {
			writeError(c, http.StatusForbidden, "forbidden_scope", "Access denied.", nil)
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func parentLinkProbeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		linkedChildID := c.Query("linked_child_id")
		if linkedChildID == "" || linkedChildID != c.Param("child_id") {
			writeError(c, http.StatusForbidden, "forbidden_parent_child_link", "Access denied.", nil)
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
