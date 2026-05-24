package application

import (
	"context"
	"testing"
	"time"

	"nursery-management-system/api/internal/modules/authentication/domain"
)

func TestRefreshUseCase_Execute(t *testing.T) {
	m1 := makeMembership(fixtureMembership1, fixtureTenantID, fixtureBranchID, "manager")
	m2 := makeMembership(fixtureMembership2, fixtureTenantID, fixtureBranchID, "practitioner")

	setup := func() (*RefreshUseCase, *fakeUserRepo, *fakeSessionRepo, *fakeTokens) {
		ur := newFakeUserRepo()
		sr := newFakeSessionRepo()
		tk := newFakeTokens()
		uc := NewRefreshUseCase(ur, sr, tk)
		return uc, ur, sr, tk
	}

	seedSession := func(sr *fakeSessionRepo, ur *fakeUserRepo) domain.RefreshToken {
		user := fixtureUser()
		ur.addUser(user)
		ur.setMemberships(user.ID, []domain.Membership{m1})

		tok := domain.RefreshToken{
			ID:           fixtureTokenID,
			UserID:       user.ID,
			MembershipID: m1.ID,
			TokenHash:    "hash:raw-refresh-token",
			ExpiresAt:    time.Now().UTC().Add(24 * time.Hour),
		}
		sr.seedToken(tok, user, m1)
		return tok
	}

	t.Run("missing refresh token returns ErrInvalidToken", func(t *testing.T) {
		uc, _, _, _ := setup()
		_, err := uc.Execute(context.Background(), "nonexistent")
		assertErrorIs(t, err, domain.ErrInvalidToken)
	})

	t.Run("successful refresh hashes raw token before lookup", func(t *testing.T) {
		uc, ur, sr, _ := setup()
		seedSession(sr, ur)

		_, err := uc.Execute(context.Background(), "raw-refresh-token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		finds := sr.callsByMethod("FindActiveRefreshToken")
		if len(finds) == 0 {
			t.Fatal("expected FindActiveRefreshToken call")
		}
		if finds[0].TokenHash != "hash:raw-refresh-token" {
			t.Fatalf("expected hash lookup, got %q", finds[0].TokenHash)
		}
	})

	t.Run("successful refresh rotates exactly one token", func(t *testing.T) {
		uc, ur, sr, _ := setup()
		old := seedSession(sr, ur)

		result, err := uc.Execute(context.Background(), "raw-refresh-token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		rotates := sr.callsByMethod("RotateRefreshToken")
		if len(rotates) != 1 {
			t.Fatalf("expected 1 rotate, got %d", len(rotates))
		}
		if rotates[0].OldTokenID != old.ID {
			t.Fatalf("expected old token ID %s, got %s", old.ID, rotates[0].OldTokenID)
		}
		if result.RefreshToken == "" {
			t.Fatal("expected replacement refresh token")
		}
	})

	t.Run("replacement keeps old token membership ID", func(t *testing.T) {
		uc, ur, sr, _ := setup()
		seedSession(sr, ur)

		_, err := uc.Execute(context.Background(), "raw-refresh-token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		rotates := sr.callsByMethod("RotateRefreshToken")
		if len(rotates) != 1 {
			t.Fatalf("expected 1 rotate")
		}
		if rotates[0].Replacement.MembershipID != m1.ID {
			t.Fatalf("expected membership %s, got %s", m1.ID, rotates[0].Replacement.MembershipID)
		}
	})

	t.Run("access token scope claims use active membership", func(t *testing.T) {
		uc, ur, sr, tk := setup()
		seedSession(sr, ur)

		_, err := uc.Execute(context.Background(), "raw-refresh-token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		scope := tk.lastScope()
		if scope.MembershipID != m1.ID.String() {
			t.Fatalf("expected scope membership %s, got %s", m1.ID, scope.MembershipID)
		}
		if scope.Role != "manager" {
			t.Fatalf("expected role manager, got %s", scope.Role)
		}
	})

	t.Run("result includes replacement token expiry memberships and active membership", func(t *testing.T) {
		uc, ur, sr, _ := setup()
		seedSession(sr, ur)

		result, err := uc.Execute(context.Background(), "raw-refresh-token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.RefreshToken == "" {
			t.Fatal("expected refresh token")
		}
		if result.RefreshExpiresAt.IsZero() {
			t.Fatal("expected refresh expiry")
		}
		if len(result.Memberships) != 1 {
			t.Fatalf("expected 1 membership, got %d", len(result.Memberships))
		}
		if result.ActiveMembership.ID != m1.ID {
			t.Fatalf("expected active %s, got %s", m1.ID, result.ActiveMembership.ID)
		}
	})

	t.Run("active membership no longer in user list returns ErrInvalidToken", func(t *testing.T) {
		uc, ur, sr, _ := setup()
		user := fixtureUser()
		ur.addUser(user)
		// User has m2 now, not m1
		ur.setMemberships(user.ID, []domain.Membership{m2})

		tok := domain.RefreshToken{
			ID:           fixtureTokenID,
			UserID:       user.ID,
			MembershipID: m1.ID,
			TokenHash:    "hash:raw-refresh-token",
			ExpiresAt:    time.Now().UTC().Add(24 * time.Hour),
		}
		sr.seedToken(tok, user, m1)

		_, err := uc.Execute(context.Background(), "raw-refresh-token")
		assertErrorIs(t, err, domain.ErrInvalidToken)
	})
}
