package application

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"nursery-management-system/api/internal/modules/authentication/domain"
)

// Deterministic UUID fixtures.
var (
	fixtureUserID         = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	fixtureTenantID       = uuid.MustParse("00000000-0000-0000-0000-000000000010")
	fixtureBranchID       = uuid.MustParse("00000000-0000-0000-0000-000000000100")
	fixtureMembership1    = uuid.MustParse("00000000-0000-0000-0000-000000001001")
	fixtureMembership2    = uuid.MustParse("00000000-0000-0000-0000-000000001002")
	fixtureMembership3    = uuid.MustParse("00000000-0000-0000-0000-000000001003")
	fixtureTokenID        = uuid.MustParse("00000000-0000-0000-0000-000000002001")
	fixtureReplaceTokenID = uuid.MustParse("00000000-0000-0000-0000-000000002002")
)

const (
	fixtureEmail    = "user@example.com"
	fixturePassword = "secret123"
)

func mustPasswordHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt hash: %v", err)
	}
	return string(hash)
}

func fixtureUser() domain.User {
	return domain.User{
		ID:           fixtureUserID,
		Email:        fixtureEmail,
		PasswordHash: mustPasswordHash(&testing.T{}, fixturePassword),
		IsActive:     true,
	}
}

func makeMembership(id, tenantID, branchID uuid.UUID, role string) domain.Membership {
	return domain.Membership{
		ID:         id,
		TenantID:   tenantID,
		TenantName: "Little Sprouts Nursery",
		BranchID:   branchID,
		BranchName: "Main Branch",
		Role:       role,
		IsActive:   true,
	}
}

func makeInactiveMembership(id, tenantID, branchID uuid.UUID, role string) domain.Membership {
	return domain.Membership{
		ID:         id,
		TenantID:   tenantID,
		TenantName: "Little Sprouts Nursery",
		BranchID:   branchID,
		BranchName: "Main Branch",
		Role:       role,
		IsActive:   false,
	}
}

// --- fakeUserRepo ---

type fakeUserRepo struct {
	mu            sync.Mutex
	userByEmail   map[string]domain.User
	memberships   map[uuid.UUID][]domain.Membership
	capturedEmail string
	findErr       error
	listErr       error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		userByEmail: make(map[string]domain.User),
		memberships: make(map[uuid.UUID][]domain.Membership),
	}
}

func (r *fakeUserRepo) FindUserByEmail(_ context.Context, email string) (domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.capturedEmail = email
	if r.findErr != nil {
		return domain.User{}, r.findErr
	}
	u, ok := r.userByEmail[email]
	if !ok {
		return domain.User{}, domain.ErrNotFound
	}
	return u, nil
}

func (r *fakeUserRepo) ListMembershipsByUserID(_ context.Context, userID uuid.UUID) ([]domain.Membership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.listErr != nil {
		return nil, r.listErr
	}
	return r.memberships[userID], nil
}

func (r *fakeUserRepo) addUser(u domain.User) {
	r.userByEmail[u.Email] = u
}

func (r *fakeUserRepo) setMemberships(userID uuid.UUID, ms []domain.Membership) {
	r.memberships[userID] = ms
}

// --- fakeSessionRepo ---

type sessionRepoCall struct {
	Method         string
	TokenHash      string
	OldTokenID     uuid.UUID
	Replacement    domain.RefreshToken
	UserAgent      string
	IPAddress      string
	ActorUserID    uuid.UUID
	FromMembership domain.Membership
	ToMembership   domain.Membership
	RequestID      string
}

type fakeSessionRepo struct {
	mu              sync.Mutex
	refreshTokens   map[string]domain.RefreshToken // hash -> token
	tokenUsers      map[uuid.UUID]domain.User
	tokenMembership map[uuid.UUID]domain.Membership

	calls     []sessionRepoCall
	createErr error
	findErr   error
	rotateErr error
	revokeErr error
	auditErr  error
}

func newFakeSessionRepo() *fakeSessionRepo {
	return &fakeSessionRepo{
		refreshTokens:   make(map[string]domain.RefreshToken),
		tokenUsers:      make(map[uuid.UUID]domain.User),
		tokenMembership: make(map[uuid.UUID]domain.Membership),
	}
}

func (r *fakeSessionRepo) CreateRefreshToken(_ context.Context, token domain.RefreshToken, userAgent, ipAddress string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, sessionRepoCall{
		Method:      "CreateRefreshToken",
		Replacement: token,
		UserAgent:   userAgent,
		IPAddress:   ipAddress,
	})
	if r.createErr != nil {
		return r.createErr
	}
	r.refreshTokens[token.TokenHash] = token
	return nil
}

func (r *fakeSessionRepo) FindActiveRefreshToken(_ context.Context, tokenHash string) (domain.RefreshToken, domain.User, domain.Membership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, sessionRepoCall{
		Method:    "FindActiveRefreshToken",
		TokenHash: tokenHash,
	})
	if r.findErr != nil {
		return domain.RefreshToken{}, domain.User{}, domain.Membership{}, r.findErr
	}
	tok, ok := r.refreshTokens[tokenHash]
	if !ok {
		return domain.RefreshToken{}, domain.User{}, domain.Membership{}, domain.ErrNotFound
	}
	user := r.tokenUsers[tok.UserID]
	membership := r.tokenMembership[tok.MembershipID]
	return tok, user, membership, nil
}

func (r *fakeSessionRepo) RotateRefreshToken(_ context.Context, oldTokenID uuid.UUID, replacement domain.RefreshToken, userAgent, ipAddress string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, sessionRepoCall{
		Method:      "RotateRefreshToken",
		OldTokenID:  oldTokenID,
		Replacement: replacement,
		UserAgent:   userAgent,
		IPAddress:   ipAddress,
	})
	if r.rotateErr != nil {
		return r.rotateErr
	}
	// Remove old token, add replacement
	for hash, tok := range r.refreshTokens {
		if tok.ID == oldTokenID {
			delete(r.refreshTokens, hash)
			break
		}
	}
	r.refreshTokens[replacement.TokenHash] = replacement
	return nil
}

func (r *fakeSessionRepo) RevokeByTokenHash(_ context.Context, tokenHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, sessionRepoCall{
		Method:    "RevokeByTokenHash",
		TokenHash: tokenHash,
	})
	if r.revokeErr != nil {
		return r.revokeErr
	}
	delete(r.refreshTokens, tokenHash)
	return nil
}

func (r *fakeSessionRepo) CreateScopeSwitchAuditLog(_ context.Context, actorUserID uuid.UUID, fromMembership, toMembership domain.Membership, requestID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, sessionRepoCall{
		Method:         "CreateScopeSwitchAuditLog",
		ActorUserID:    actorUserID,
		FromMembership: fromMembership,
		ToMembership:   toMembership,
		RequestID:      requestID,
	})
	if r.auditErr != nil {
		return r.auditErr
	}
	return nil
}

func (r *fakeSessionRepo) seedToken(token domain.RefreshToken, user domain.User, membership domain.Membership) {
	r.refreshTokens[token.TokenHash] = token
	r.tokenUsers[user.ID] = user
	r.tokenMembership[membership.ID] = membership
}

func (r *fakeSessionRepo) callsByMethod(method string) []sessionRepoCall {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []sessionRepoCall
	for _, c := range r.calls {
		if c.Method == method {
			out = append(out, c)
		}
	}
	return out
}

// --- fakeTokens ---

const fakeAccessToken = "fake-access-token"

type fakeTokens struct {
	mu        sync.Mutex
	accessTTL int64
	generated []domain.ScopeClaims
	hashFunc  func(raw string) string
	genErr    error
}

func newFakeTokens() *fakeTokens {
	return &fakeTokens{
		accessTTL: 900,
		hashFunc:  func(raw string) string { return "hash:" + raw },
	}
}

func (f *fakeTokens) GenerateAccessToken(userID uuid.UUID, email string, scope domain.ScopeClaims) (string, time.Time, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.genErr != nil {
		return "", time.Time{}, f.genErr
	}
	f.generated = append(f.generated, scope)
	return fakeAccessToken, time.Now().UTC().Add(15 * time.Minute), nil
}

func (f *fakeTokens) GenerateRefreshToken(rememberMe bool) (raw string, hash string, expiresAt time.Time, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.genErr != nil {
		return "", "", time.Time{}, f.genErr
	}
	raw = "raw-refresh-token"
	hash = f.hashFunc(raw)
	if rememberMe {
		return raw, hash, time.Now().UTC().Add(30 * 24 * time.Hour), nil
	}
	return raw, hash, time.Now().UTC().Add(24 * time.Hour), nil
}

func (f *fakeTokens) HashRefreshToken(raw string) string {
	return f.hashFunc(raw)
}

func (f *fakeTokens) AccessTTLSeconds() int64 {
	return f.accessTTL
}

func (f *fakeTokens) lastScope() domain.ScopeClaims {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.generated) == 0 {
		return domain.ScopeClaims{}
	}
	return f.generated[len(f.generated)-1]
}
