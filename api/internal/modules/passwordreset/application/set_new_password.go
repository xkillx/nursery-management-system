package application

import (
	"context"
	"log/slog"

	"golang.org/x/crypto/bcrypt"

	"nursery-management-system/api/internal/modules/passwordreset/domain"
	domainerrors "nursery-management-system/api/internal/platform/errors"
)

type SetNewPasswordUseCase struct {
	repo   domain.Repository
	logger *slog.Logger
}

func NewSetNewPasswordUseCase(
	repo domain.Repository,
	logger *slog.Logger,
) *SetNewPasswordUseCase {
	return &SetNewPasswordUseCase{
		repo:   repo,
		logger: logger,
	}
}

type SetNewPasswordResult struct{}

func (uc *SetNewPasswordUseCase) Execute(ctx context.Context, token string, newPassword string) (SetNewPasswordResult, error) {
	if len(newPassword) < 8 {
		return SetNewPasswordResult{}, domainerrors.Validation("Password must be at least 8 characters.", "new_password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return SetNewPasswordResult{}, domainerrors.Internal(err)
	}

	err = uc.repo.ResetPassword(ctx, token, string(hash))
	if err != nil {
		switch {
		case err == domain.ErrTokenInvalid:
			return SetNewPasswordResult{}, domainerrors.New("password_reset_token_invalid", "Invalid reset token.")
		case err == domain.ErrTokenExpired:
			return SetNewPasswordResult{}, domainerrors.New("password_reset_token_expired", "Reset token has expired.")
		case err == domain.ErrTokenUsed:
			return SetNewPasswordResult{}, domainerrors.New("password_reset_token_used", "Reset token has already been used.")
		default:
			return SetNewPasswordResult{}, domainerrors.Internal(err)
		}
	}

	uc.logger.Info("password_reset_completed")
	return SetNewPasswordResult{}, nil
}
