package application

import (
	"context"

	"nursery-management-system/api/internal/modules/authentication/domain"
)

type LogoutUseCase struct {
	sessionRepo domain.SessionRepository
	tokens      TokenGenerator
}

func NewLogoutUseCase(sessionRepo domain.SessionRepository, tokens TokenGenerator) *LogoutUseCase {
	return &LogoutUseCase{
		sessionRepo: sessionRepo,
		tokens:      tokens,
	}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, rawRefreshToken string) error {
	refreshHash := uc.tokens.HashRefreshToken(rawRefreshToken)
	_ = uc.sessionRepo.RevokeByTokenHash(ctx, refreshHash)
	return nil
}
