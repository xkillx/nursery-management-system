package application

import (
	"context"
	"time"

	"nursery-management-system/api/internal/modules/authentication/domain"
	"nursery-management-system/api/internal/platform/uid"
)

type SwitchResult struct {
	User             domain.User
	Memberships      []domain.Membership
	ActiveMembership domain.Membership
	AccessToken      string
	RefreshToken     string
	RefreshExpiresAt time.Time
}

type SwitchMembershipUseCase struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
	tokens      TokenGenerator
}

func NewSwitchMembershipUseCase(userRepo domain.UserRepository, sessionRepo domain.SessionRepository, tokens TokenGenerator) *SwitchMembershipUseCase {
	return &SwitchMembershipUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		tokens:      tokens,
	}
}

func (uc *SwitchMembershipUseCase) Execute(ctx context.Context, rawRefreshToken, targetMembershipID, requestID string) (SwitchResult, error) {
	refreshHash := uc.tokens.HashRefreshToken(rawRefreshToken)

	oldToken, user, oldMembership, err := uc.sessionRepo.FindActiveRefreshToken(ctx, refreshHash)
	if err != nil {
		return SwitchResult{}, domain.ErrInvalidToken
	}

	memberships, err := uc.userRepo.ListMembershipsByUserID(ctx, user.ID)
	if err != nil {
		return SwitchResult{}, err
	}

	selectedMembership, err := SelectExplicitMembership(memberships, targetMembershipID)
	if err != nil {
		return SwitchResult{}, err
	}

	rawReplacement, replacementHash, replacementExpiresAt, err := uc.tokens.GenerateRefreshToken(oldToken.RememberMe)
	if err != nil {
		return SwitchResult{}, err
	}

	replacement := domain.RefreshToken{
		ID:           uid.NewUUID(),
		UserID:       user.ID,
		MembershipID: selectedMembership.ID,
		TokenHash:    replacementHash,
		ExpiresAt:    replacementExpiresAt,
		RememberMe:   oldToken.RememberMe,
	}

	if err := uc.sessionRepo.RotateRefreshToken(ctx, oldToken.ID, replacement, userAgentFromContext(ctx), ipAddressFromContext(ctx)); err != nil {
		return SwitchResult{}, err
	}

	if selectedMembership.ID != oldMembership.ID {
		if err := uc.sessionRepo.CreateScopeSwitchAuditLog(ctx, user.ID, oldMembership, selectedMembership, requestID); err != nil {
			return SwitchResult{}, err
		}
	}

	accessToken, _, err := uc.tokens.GenerateAccessToken(user.ID, user.Email, domain.ScopeClaims{
		MembershipID: selectedMembership.ID.String(),
		TenantID:     selectedMembership.TenantID.String(),
		BranchID:     scopeBranchID(selectedMembership),
		Role:         selectedMembership.Role,
	})
	if err != nil {
		return SwitchResult{}, err
	}

	return SwitchResult{
		User:             user,
		Memberships:      memberships,
		ActiveMembership: selectedMembership,
		AccessToken:      accessToken,
		RefreshToken:     rawReplacement,
		RefreshExpiresAt: replacementExpiresAt,
	}, nil
}
