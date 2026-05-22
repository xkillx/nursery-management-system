package tenant

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const AuthContextKey = "auth_context"

type AuthorizationContext struct {
	UserID       string `json:"user_id"`
	MembershipID string `json:"membership_id"`
	TenantID     string `json:"tenant_id"`
	BranchID     string `json:"branch_id"`
	Role         string `json:"role"`
	RequestID    string `json:"request_id"`
}

type ActorContext struct {
	UserID       uuid.UUID
	MembershipID uuid.UUID
	TenantID     uuid.UUID
	BranchID     uuid.UUID
	RequestID    string
}

func ActorFromGinContext(c *gin.Context) (ActorContext, bool) {
	v, ok := c.Get(AuthContextKey)
	if !ok {
		return ActorContext{}, false
	}

	authCtx, ok := v.(AuthorizationContext)
	if !ok {
		return ActorContext{}, false
	}

	userID, err := uuid.Parse(authCtx.UserID)
	if err != nil {
		return ActorContext{}, false
	}
	membershipID, err := uuid.Parse(authCtx.MembershipID)
	if err != nil {
		return ActorContext{}, false
	}
	tenantID, err := uuid.Parse(authCtx.TenantID)
	if err != nil {
		return ActorContext{}, false
	}
	branchID, err := uuid.Parse(authCtx.BranchID)
	if err != nil {
		return ActorContext{}, false
	}

	return ActorContext{
		UserID:       userID,
		MembershipID: membershipID,
		TenantID:     tenantID,
		BranchID:     branchID,
		RequestID:    authCtx.RequestID,
	}, true
}
