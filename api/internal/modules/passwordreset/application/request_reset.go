package application

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"nursery-management-system/api/internal/modules/passwordreset/domain"
)

type TokenGenerator interface {
	Generate() (raw string, hash string, expiresAt time.Time, err error)
}

type RequestResetUseCase struct {
	repo       domain.Repository
	email      domain.EmailSender
	tokens     TokenGenerator
	webBaseURL string
	logger     *slog.Logger
}

func NewRequestResetUseCase(
	repo domain.Repository,
	email domain.EmailSender,
	tokens TokenGenerator,
	webBaseURL string,
	logger *slog.Logger,
) *RequestResetUseCase {
	return &RequestResetUseCase{
		repo:       repo,
		email:      email,
		tokens:     tokens,
		webBaseURL: webBaseURL,
		logger:     logger,
	}
}

type RequestResetResult struct {
	Accepted bool
}

func (uc *RequestResetUseCase) Execute(ctx context.Context, email string) (RequestResetResult, error) {
	emailNormalized := strings.ToLower(strings.TrimSpace(email))

	user, found, err := uc.repo.FindUserByEmail(ctx, emailNormalized)
	if err != nil {
		return RequestResetResult{}, err
	}

	if !found || !user.IsActive {
		uc.logger.Info("password_reset_request_accepted_silent", "email_normalized", emailNormalized)
		return RequestResetResult{Accepted: true}, nil
	}

	raw, hash, expiresAt, err := uc.tokens.Generate()
	if err != nil {
		return RequestResetResult{}, err
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", uc.webBaseURL, url.QueryEscape(raw))

	err = uc.repo.IssueResetToken(ctx, user.ID, hash, expiresAt, func() error {
		return uc.email.SendPasswordReset(ctx, user.Email, resetURL)
	})
	if err != nil {
		uc.logger.Error("password_reset_issue_failed", "user_id", user.ID, "error", err)
		return RequestResetResult{}, err
	}

	uc.logger.Info("password_reset_issued", "user_id", user.ID)
	return RequestResetResult{Accepted: true}, nil
}
