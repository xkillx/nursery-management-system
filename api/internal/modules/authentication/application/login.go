package application

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"nursery-management-system/api/internal/modules/authentication/domain"
	"nursery-management-system/api/internal/platform/uid"
)

type LoginResult struct {
	User             domain.User
	Memberships      []domain.Membership
	ActiveMembership domain.Membership
	AccessToken      string
	RefreshToken     string
	RefreshExpiresAt time.Time
}

type LoginUseCase struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
	tokens      TokenGenerator
}

type TokenGenerator interface {
	GenerateAccessToken(userID uuid.UUID, email string, scope domain.ScopeClaims) (string, time.Time, error)
	GenerateRefreshToken() (raw string, hash string, expiresAt time.Time, err error)
	HashRefreshToken(raw string) string
	AccessTTLSeconds() int64
}

func NewLoginUseCase(userRepo domain.UserRepository, sessionRepo domain.SessionRepository, tokens TokenGenerator) *LoginUseCase {
	return &LoginUseCase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		tokens:      tokens,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, email, password, membershipID string) (LoginResult, error) {
	emailNormalized := strings.ToLower(strings.TrimSpace(email))

	user, err := uc.userRepo.FindUserByEmail(ctx, emailNormalized)
	if err != nil || !user.IsActive {
		return LoginResult{}, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return LoginResult{}, domain.ErrInvalidCredentials
	}

	memberships, err := uc.userRepo.ListMembershipsByUserID(ctx, user.ID)
	if err != nil {
		return LoginResult{}, err
	}

	activeMembership, err := SelectLoginMembership(memberships, membershipID)
	if err != nil {
		return LoginResult{}, err
	}

	accessToken, _, err := uc.tokens.GenerateAccessToken(user.ID, user.Email, domain.ScopeClaims{
		MembershipID: activeMembership.ID.String(),
		TenantID:     activeMembership.TenantID.String(),
		BranchID:     activeMembership.BranchID.String(),
		Role:         activeMembership.Role,
	})
	if err != nil {
		return LoginResult{}, err
	}

	rawRefresh, refreshHash, refreshExpiresAt, err := uc.tokens.GenerateRefreshToken()
	if err != nil {
		return LoginResult{}, err
	}

	err = uc.sessionRepo.CreateRefreshToken(ctx, domain.RefreshToken{
		ID:           uid.NewUUID(),
		UserID:       user.ID,
		MembershipID: activeMembership.ID,
		TokenHash:    refreshHash,
		ExpiresAt:    refreshExpiresAt,
	}, userAgentFromContext(ctx), ipAddressFromContext(ctx))
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		User:             user,
		Memberships:      memberships,
		ActiveMembership: activeMembership,
		AccessToken:      accessToken,
		RefreshToken:     rawRefresh,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

type contextKey string

const (
	userAgentKey contextKey = "user_agent"
	ipAddressKey contextKey = "ip_address"
)

func ContextWithRequestMeta(ctx context.Context, userAgent, ipAddress string) context.Context {
	ctx = context.WithValue(ctx, userAgentKey, userAgent)
	ctx = context.WithValue(ctx, ipAddressKey, ipAddress)
	return ctx
}

func userAgentFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(userAgentKey).(string); ok {
		return v
	}
	return ""
}

func ipAddressFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ipAddressKey).(string); ok {
		return v
	}
	return ""
}
