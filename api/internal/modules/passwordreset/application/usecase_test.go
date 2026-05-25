package application

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"nursery-management-system/api/internal/modules/passwordreset/domain"
)

type mockRepo struct {
	user          domain.User
	userFound     bool
	userErr       error
	issueErr      error
	resetErr      error
	issuedTokens  []string
	emailCaptured string
}

func (m *mockRepo) FindUserByEmail(_ context.Context, _ string) (domain.User, bool, error) {
	return m.user, m.userFound, m.userErr
}

func (m *mockRepo) IssueResetToken(_ context.Context, _ uuid.UUID, tokenHash string, _ time.Time, sendEmail func() error) error {
	m.issuedTokens = append(m.issuedTokens, tokenHash)
	if err := sendEmail(); err != nil {
		return err
	}
	if m.issueErr != nil {
		return m.issueErr
	}
	return nil
}

func (m *mockRepo) ResetPassword(_ context.Context, _ string, _ string) error {
	return m.resetErr
}

type mockEmailSender struct {
	captured []string
	err      error
}

func (m *mockEmailSender) SendPasswordReset(_ context.Context, to string, resetURL string) error {
	m.captured = append(m.captured, to+"|"+resetURL)
	if m.err != nil {
		return m.err
	}
	return nil
}

type mockTokenGen struct {
	raw      string
	hash     string
	expiresAt time.Time
	err      error
}

func (m *mockTokenGen) Generate() (string, string, time.Time, error) {
	return m.raw, m.hash, m.expiresAt, m.err
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestRequestReset_NormalizesEmail(t *testing.T) {
	repo := &mockRepo{user: domain.User{ID: uuid.New(), Email: "test@example.com", IsActive: true}, userFound: true}
	email := &mockEmailSender{}
	tok := &mockTokenGen{raw: "abc", hash: "hash123", expiresAt: time.Now().Add(60 * time.Minute)}

	uc := NewRequestResetUseCase(repo, email, tok, "http://localhost:4200", testLogger())
	_, err := uc.Execute(context.Background(), " Test@Example.COM ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(email.captured) != 1 {
		t.Fatal("expected email to be sent")
	}
}

func TestRequestReset_UnknownUser_ReturnsAccepted(t *testing.T) {
	repo := &mockRepo{userFound: false}
	email := &mockEmailSender{}
	tok := &mockTokenGen{}

	uc := NewRequestResetUseCase(repo, email, tok, "http://localhost:4200", testLogger())
	result, err := uc.Execute(context.Background(), "unknown@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Accepted {
		t.Fatal("expected accepted")
	}
	if len(email.captured) != 0 {
		t.Fatal("expected no email for unknown user")
	}
	if len(repo.issuedTokens) != 0 {
		t.Fatal("expected no token for unknown user")
	}
}

func TestRequestReset_InactiveUser_ReturnsAccepted(t *testing.T) {
	repo := &mockRepo{user: domain.User{ID: uuid.New(), IsActive: false}, userFound: true}
	email := &mockEmailSender{}
	tok := &mockTokenGen{}

	uc := NewRequestResetUseCase(repo, email, tok, "http://localhost:4200", testLogger())
	result, err := uc.Execute(context.Background(), "inactive@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Accepted {
		t.Fatal("expected accepted")
	}
	if len(email.captured) != 0 {
		t.Fatal("expected no email for inactive user")
	}
}

func TestRequestReset_ActiveUser_SendsEmail(t *testing.T) {
	repo := &mockRepo{user: domain.User{ID: uuid.New(), Email: "active@example.com", IsActive: true}, userFound: true}
	email := &mockEmailSender{}
	tok := &mockTokenGen{raw: "raw-token", hash: "hashed", expiresAt: time.Now().Add(60 * time.Minute)}

	uc := NewRequestResetUseCase(repo, email, tok, "http://localhost:4200", testLogger())
	result, err := uc.Execute(context.Background(), "active@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Accepted {
		t.Fatal("expected accepted")
	}
	if len(email.captured) != 1 {
		t.Fatal("expected one email sent")
	}
	if !strings.Contains(email.captured[0], "raw-token") {
		t.Fatal("expected reset URL to contain raw token")
	}
}

func TestRequestReset_EmailFailure_ReturnsError(t *testing.T) {
	repo := &mockRepo{user: domain.User{ID: uuid.New(), Email: "active@example.com", IsActive: true}, userFound: true}
	email := &mockEmailSender{err: errors.New("smtp down")}
	tok := &mockTokenGen{raw: "raw", hash: "hash", expiresAt: time.Now().Add(60 * time.Minute)}

	uc := NewRequestResetUseCase(repo, email, tok, "http://localhost:4200", testLogger())
	_, err := uc.Execute(context.Background(), "active@example.com")
	if err == nil {
		t.Fatal("expected error on email failure")
	}
}

func TestSetNewPassword_ShortPassword_ReturnsValidationError(t *testing.T) {
	repo := &mockRepo{}
	uc := NewSetNewPasswordUseCase(repo, testLogger())
	_, err := uc.Execute(context.Background(), "token", "short")
	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Error() == "" {
		t.Fatal("expected error message")
	}
}

func TestSetNewPassword_InvalidToken_ReturnsTokenInvalidError(t *testing.T) {
	repo := &mockRepo{resetErr: domain.ErrTokenInvalid}
	uc := NewSetNewPasswordUseCase(repo, testLogger())
	_, err := uc.Execute(context.Background(), "bad-token", "newpassword123")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "password_reset_token_invalid") {
		t.Fatalf("expected token invalid error, got: %v", err)
	}
}

func TestSetNewPassword_ExpiredToken_ReturnsTokenExpiredError(t *testing.T) {
	repo := &mockRepo{resetErr: domain.ErrTokenExpired}
	uc := NewSetNewPasswordUseCase(repo, testLogger())
	_, err := uc.Execute(context.Background(), "expired-token", "newpassword123")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "password_reset_token_expired") {
		t.Fatalf("expected token expired error, got: %v", err)
	}
}

func TestSetNewPassword_UsedToken_ReturnsTokenUsedError(t *testing.T) {
	repo := &mockRepo{resetErr: domain.ErrTokenUsed}
	uc := NewSetNewPasswordUseCase(repo, testLogger())
	_, err := uc.Execute(context.Background(), "used-token", "newpassword123")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "password_reset_token_used") {
		t.Fatalf("expected token used error, got: %v", err)
	}
}

func TestSetNewPassword_ValidToken_ReturnsSuccess(t *testing.T) {
	repo := &mockRepo{}
	uc := NewSetNewPasswordUseCase(repo, testLogger())
	_, err := uc.Execute(context.Background(), "valid-token", "newpassword123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetNewPassword_HashesPassword(t *testing.T) {
	var capturedHash string
	repo := &mockRepo{}
	repo.resetErr = nil
	repoImpl := &capturingRepo{inner: repo, capturedHash: &capturedHash}

	uc := NewSetNewPasswordUseCase(repoImpl, testLogger())
	_, err := uc.Execute(context.Background(), "valid-token", "newpassword123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedHash == "" {
		t.Fatal("expected password hash to be set")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(capturedHash), []byte("newpassword123")); err != nil {
		t.Fatal("expected hash to match new password")
	}
}

type capturingRepo struct {
	inner        *mockRepo
	capturedHash *string
}

func (c *capturingRepo) FindUserByEmail(ctx context.Context, email string) (domain.User, bool, error) {
	return c.inner.FindUserByEmail(ctx, email)
}

func (c *capturingRepo) IssueResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time, sendEmail func() error) error {
	return c.inner.IssueResetToken(ctx, userID, tokenHash, expiresAt, sendEmail)
}

func (c *capturingRepo) ResetPassword(_ context.Context, _ string, newPasswordHash string) error {
	*c.capturedHash = newPasswordHash
	return nil
}
