package application

import (
	"context"
	"testing"
	"time"

	"nursery-management-system/api/internal/modules/authentication/domain"
)

func TestSwitchMembershipUseCase_Execute(t *testing.T) {
	m1 := makeMembership(fixtureMembership1, fixtureTenantID, fixtureBranchID, "manager")
	m2 := makeMembership(fixtureMembership2, fixtureTenantID, fixtureBranchID, "practitioner")

	setup := func() (*SwitchMembershipUseCase, *fakeUserRepo, *fakeSessionRepo, *fakeTokens) {
		ur := newFakeUserRepo()
		sr := newFakeSessionRepo()
		tk := newFakeTokens()
		uc := NewSwitchMembershipUseCase(ur, sr, tk)
		return uc, ur, sr, tk
	}

	seedSession := func(sr *fakeSessionRepo, ur *fakeUserRepo, membership domain.Membership) domain.RefreshToken {
		user := fixtureUser()
		ur.addUser(user)
		ur.setMemberships(user.ID, []domain.Membership{m1, m2})

		tok := domain.RefreshToken{
			ID:           fixtureTokenID,
			UserID:       user.ID,
			MembershipID: membership.ID,
			TokenHash:    "hash:raw-refresh-token",
			ExpiresAt:    time.Now().UTC().Add(24 * time.Hour),
		}
		sr.seedToken(tok, user, membership)
		return tok
	}

	t.Run("missing refresh token returns ErrInvalidToken", func(t *testing.T) {
		uc, _, _, _ := setup()
		_, err := uc.Execute(context.Background(), "nonexistent", fixtureMembership2.String(), "req-1")
		assertErrorIs(t, err, domain.ErrInvalidToken)
	})

	t.Run("malformed target membership returns ErrInvalidMembership and no rotation", func(t *testing.T) {
		uc, ur, sr, _ := setup()
		seedSession(sr, ur, m1)

		_, err := uc.Execute(context.Background(), "raw-refresh-token", "bad-uuid", "req-1")
		assertErrorIs(t, err, domain.ErrInvalidMembership)
		rotates := sr.callsByMethod("RotateRefreshToken")
		if len(rotates) != 0 {
			t.Fatalf("expected 0 rotations, got %d", len(rotates))
		}
	})

	t.Run("unknown target membership returns ErrInvalidMembership and no rotation", func(t *testing.T) {
		uc, ur, sr, _ := setup()
		seedSession(sr, ur, m1)

		_, err := uc.Execute(context.Background(), "raw-refresh-token", fixtureMembership3.String(), "req-1")
		assertErrorIs(t, err, domain.ErrInvalidMembership)
		rotates := sr.callsByMethod("RotateRefreshToken")
		if len(rotates) != 0 {
			t.Fatalf("expected 0 rotations, got %d", len(rotates))
		}
	})

	t.Run("switching to different active membership rotates to target", func(t *testing.T) {
		uc, ur, sr, _ := setup()
		seedSession(sr, ur, m1)

		result, err := uc.Execute(context.Background(), "raw-refresh-token", fixtureMembership2.String(), "req-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		rotates := sr.callsByMethod("RotateRefreshToken")
		if len(rotates) != 1 {
			t.Fatalf("expected 1 rotate, got %d", len(rotates))
		}
		if rotates[0].Replacement.MembershipID != m2.ID {
			t.Fatalf("expected replacement membership %s, got %s", m2.ID, rotates[0].Replacement.MembershipID)
		}
		if result.ActiveMembership.ID != m2.ID {
			t.Fatalf("expected active %s, got %s", m2.ID, result.ActiveMembership.ID)
		}
	})

	t.Run("switching to different membership emits one audit call", func(t *testing.T) {
		uc, ur, sr, _ := setup()
		seedSession(sr, ur, m1)

		_, err := uc.Execute(context.Background(), "raw-refresh-token", fixtureMembership2.String(), "req-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		audits := sr.callsByMethod("CreateScopeSwitchAuditLog")
		if len(audits) != 1 {
			t.Fatalf("expected 1 audit, got %d", len(audits))
		}
		a := audits[0]
		if a.ActorUserID != fixtureUserID {
			t.Fatalf("expected actor %s, got %s", fixtureUserID, a.ActorUserID)
		}
		if a.FromMembership.ID != m1.ID {
			t.Fatalf("expected from %s, got %s", m1.ID, a.FromMembership.ID)
		}
		if a.ToMembership.ID != m2.ID {
			t.Fatalf("expected to %s, got %s", m2.ID, a.ToMembership.ID)
		}
		if a.RequestID != "req-123" {
			t.Fatalf("expected requestID req-123, got %s", a.RequestID)
		}
	})

	t.Run("switching to same membership rotates but no audit", func(t *testing.T) {
		uc, ur, sr, _ := setup()
		seedSession(sr, ur, m1)

		_, err := uc.Execute(context.Background(), "raw-refresh-token", fixtureMembership1.String(), "req-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		rotates := sr.callsByMethod("RotateRefreshToken")
		if len(rotates) != 1 {
			t.Fatalf("expected 1 rotate, got %d", len(rotates))
		}
		audits := sr.callsByMethod("CreateScopeSwitchAuditLog")
		if len(audits) != 0 {
			t.Fatalf("expected 0 audits, got %d", len(audits))
		}
	})

	t.Run("access token scope claims match target membership", func(t *testing.T) {
		uc, ur, sr, tk := setup()
		seedSession(sr, ur, m1)

		_, err := uc.Execute(context.Background(), "raw-refresh-token", fixtureMembership2.String(), "req-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		scope := tk.lastScope()
		if scope.MembershipID != fixtureMembership2.String() {
			t.Fatalf("expected scope membership %s, got %s", fixtureMembership2, scope.MembershipID)
		}
		if scope.Role != "practitioner" {
			t.Fatalf("expected role practitioner, got %s", scope.Role)
		}
	})

	t.Run("result includes all memberships and target active membership", func(t *testing.T) {
		uc, ur, sr, _ := setup()
		seedSession(sr, ur, m1)

		result, err := uc.Execute(context.Background(), "raw-refresh-token", fixtureMembership2.String(), "req-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Memberships) != 2 {
			t.Fatalf("expected 2 memberships, got %d", len(result.Memberships))
		}
		if result.ActiveMembership.ID != m2.ID {
			t.Fatalf("expected active %s, got %s", m2.ID, result.ActiveMembership.ID)
		}
	})
}
