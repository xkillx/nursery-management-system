package application

import (
	"context"
	"time"

	"nursery-management-system/api/internal/modules/authentication/domain"
	"nursery-management-system/api/internal/platform/uid"
)

type RefreshResult struct {
	User             domain.User
	Memberships      []domain.Membership
	ActiveMembership domain.Membership
	AccessToken      string
	RefreshToken     string
	RefreshExpiresAt time.Time
}

type RefreshUseCase struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
	tokens      TokenGenerator
}

func NewRefreshUseCase(userRepo domain.UserRepository, sessionRepo domain.SessionRepository, tokens TokenGenerator) *RefreshUseCase {
	return &RefreshUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		tokens:      tokens,
	}
}

func (uc *RefreshUseCase) Execute(ctx context.Context, rawRefreshToken string) (RefreshResult, error) {
	refreshHash := uc.tokens.HashRefreshToken(rawRefreshToken)

	oldToken, user, activeMembership, err := uc.sessionRepo.FindActiveRefreshToken(ctx, refreshHash)
	if err != nil {
		return RefreshResult{}, domain.ErrInvalidToken
	}

	rawReplacement, replacementHash, replacementExpiresAt, err := uc.tokens.GenerateRefreshToken()
	if err != nil {
		return RefreshResult{}, err
	}

	replacement := domain.RefreshToken{
		ID:           uid.NewUUID(),
		UserID:       user.ID,
		MembershipID: oldToken.MembershipID,
		TokenHash:    replacementHash,
		ExpiresAt:    replacementExpiresAt,
	}

	if err := uc.sessionRepo.RotateRefreshToken(ctx, oldToken.ID, replacement, userAgentFromContext(ctx), ipAddressFromContext(ctx)); err != nil {
		return RefreshResult{}, err
	}

	memberships, err := uc.userRepo.ListMembershipsByUserID(ctx, user.ID)
	if err != nil {
		return RefreshResult{}, err
	}

	if !containsMembership(memberships, activeMembership.ID) {
		return RefreshResult{}, domain.ErrInvalidToken
	}

	accessToken, _, err := uc.tokens.GenerateAccessToken(user.ID, user.Email, domain.ScopeClaims{
		MembershipID: activeMembership.ID.String(),
		TenantID:     activeMembership.TenantID.String(),
		BranchID:     activeMembership.BranchID.String(),
		Role:         activeMembership.Role,
	})
	if err != nil {
		return RefreshResult{}, err
	}

	return RefreshResult{
		User:             user,
		Memberships:      memberships,
		ActiveMembership: activeMembership,
		AccessToken:      accessToken,
		RefreshToken:     rawReplacement,
		RefreshExpiresAt: replacementExpiresAt,
	}, nil
}
