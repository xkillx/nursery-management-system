package bootstrap

import (
	"nursery-management-system/api/internal/modules/authentication/infrastructure/tokens"
	"nursery-management-system/api/internal/platform/tenant"
)

type tokenParserAdapter struct {
	tm *tokens.TokenManager
}

func (a *tokenParserAdapter) ParseAccessToken(raw string) (tenant.AuthorizationContext, error) {
	claims, err := a.tm.ParseAccessToken(raw)
	if err != nil {
		return tenant.AuthorizationContext{}, err
	}
	return tenant.AuthorizationContext{
		UserID:       claims.Subject,
		MembershipID: claims.MembershipID,
		TenantID:     claims.TenantID,
		BranchID:     claims.BranchID,
		Role:         claims.Role,
	}, nil
}
