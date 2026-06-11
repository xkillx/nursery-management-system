package tenant

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const AuthContextKey = "auth_context"

type AuthorizationContext struct {
	UserID        string `json:"user_id"`
	MembershipID  string `json:"membership_id"`
	TenantID      string `json:"tenant_id"`
	BranchID      string `json:"branch_id"`
	Role          string `json:"role"`
	RequestID     string `json:"request_id"`
	CorrelationID string `json:"correlation_id"`
	TraceID       string `json:"trace_id"`
}

type ActorContext struct {
	UserID        uuid.UUID
	MembershipID  uuid.UUID
	TenantID      uuid.UUID
	BranchID      uuid.UUID
	RequestID     string
	CorrelationID string
	TraceID       string
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
	if err != nil || branchID == uuid.Nil {
		return ActorContext{}, false
	}

	return ActorContext{
		UserID:        userID,
		MembershipID:  membershipID,
		TenantID:      tenantID,
		BranchID:      branchID,
		RequestID:     authCtx.RequestID,
		CorrelationID: authCtx.CorrelationID,
		TraceID:       authCtx.TraceID,
	}, true
}

type OwnerActorContext struct {
	UserID        uuid.UUID
	MembershipID  uuid.UUID
	TenantID      uuid.UUID
	RequestID     string
	CorrelationID string
	TraceID       string
}

func OwnerActorFromGinContext(c *gin.Context) (OwnerActorContext, bool) {
	v, ok := c.Get(AuthContextKey)
	if !ok {
		return OwnerActorContext{}, false
	}

	authCtx, ok := v.(AuthorizationContext)
	if !ok {
		return OwnerActorContext{}, false
	}

	if authCtx.Role != "owner" {
		return OwnerActorContext{}, false
	}

	userID, err := uuid.Parse(authCtx.UserID)
	if err != nil {
		return OwnerActorContext{}, false
	}
	membershipID, err := uuid.Parse(authCtx.MembershipID)
	if err != nil {
		return OwnerActorContext{}, false
	}
	tenantID, err := uuid.Parse(authCtx.TenantID)
	if err != nil {
		return OwnerActorContext{}, false
	}

	return OwnerActorContext{
		UserID:        userID,
		MembershipID:  membershipID,
		TenantID:      tenantID,
		RequestID:     authCtx.RequestID,
		CorrelationID: authCtx.CorrelationID,
		TraceID:       authCtx.TraceID,
	}, true
}
