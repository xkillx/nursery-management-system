package application

import (
	"context"
	"testing"

	"nursery-management-system/api/internal/modules/authentication/domain"
)

func TestLoginUseCase_Execute(t *testing.T) {
	m1 := makeMembership(fixtureMembership1, fixtureTenantID, fixtureBranchID, "manager")
	m2 := makeMembership(fixtureMembership2, fixtureTenantID, fixtureBranchID, "practitioner")

	setup := func() (*LoginUseCase, *fakeUserRepo, *fakeSessionRepo, *fakeTokens) {
		ur := newFakeUserRepo()
		sr := newFakeSessionRepo()
		tk := newFakeTokens()
		uc := NewLoginUseCase(ur, sr, tk)
		return uc, ur, sr, tk
	}

	t.Run("email lowercased and trimmed before lookup", func(t *testing.T) {
		uc, ur, _, _ := setup()
		ur.addUser(fixtureUser())
		ur.setMemberships(fixtureUserID, []domain.Membership{m1})

		_, err := uc.Execute(context.Background(), " User@Example.com ", fixturePassword, "", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ur.capturedEmail != "user@example.com" {
			t.Fatalf("expected normalized email user@example.com, got %q", ur.capturedEmail)
		}
	})

	t.Run("unknown user returns ErrInvalidCredentials", func(t *testing.T) {
		uc, _, _, _ := setup()
		_, err := uc.Execute(context.Background(), "nobody@example.com", fixturePassword, "", true)
		assertErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("inactive user returns ErrInvalidCredentials", func(t *testing.T) {
		uc, ur, _, _ := setup()
		inactive := fixtureUser()
		inactive.IsActive = false
		ur.addUser(inactive)

		_, err := uc.Execute(context.Background(), fixtureEmail, fixturePassword, "", true)
		assertErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("bad password returns ErrInvalidCredentials", func(t *testing.T) {
		uc, ur, _, _ := setup()
		ur.addUser(fixtureUser())

		_, err := uc.Execute(context.Background(), fixtureEmail, "wrongpassword", "", true)
		assertErrorIs(t, err, domain.ErrInvalidCredentials)
	})

	t.Run("multi-membership without selection returns MembershipSelectionRequiredError and no refresh token", func(t *testing.T) {
		uc, ur, sr, _ := setup()
		ur.addUser(fixtureUser())
		ur.setMemberships(fixtureUserID, []domain.Membership{m1, m2})

		_, err := uc.Execute(context.Background(), fixtureEmail, fixturePassword, "", true)
		var selErr *domain.MembershipSelectionRequiredError
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !isMembershipSelectionRequired(err) {
			t.Fatalf("expected MembershipSelectionRequiredError, got %T: %v", err, err)
		}
		selErr = err.(*domain.MembershipSelectionRequiredError)
		if selErr.IsStaleChoice {
			t.Fatal("expected IsStaleChoice=false for initial challenge")
		}
		if len(selErr.Memberships) != 2 {
			t.Fatalf("expected 2 choices, got %d", len(selErr.Memberships))
		}
		creates := sr.callsByMethod("CreateRefreshToken")
		if len(creates) != 0 {
			t.Fatalf("expected 0 creates, got %d", len(creates))
		}
	})

	t.Run("multi-membership malformed selection returns validation error and no refresh token", func(t *testing.T) {
		uc, ur, sr, _ := setup()
		ur.addUser(fixtureUser())
		ur.setMemberships(fixtureUserID, []domain.Membership{m1, m2})

		_, err := uc.Execute(context.Background(), fixtureEmail, fixturePassword, "not-a-uuid", true)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var valErr *domain.ValidationError
		if !isValidationError(err) {
			t.Fatalf("expected ValidationError, got %T: %v", err, err)
		}
		valErr = err.(*domain.ValidationError)
		if valErr.Field != "membership_id" {
			t.Fatalf("expected field membership_id, got %s", valErr.Field)
		}
		creates := sr.callsByMethod("CreateRefreshToken")
		if len(creates) != 0 {
			t.Fatalf("expected 0 creates, got %d", len(creates))
		}
	})

	t.Run("multi-membership stale selection returns MembershipSelectionRequiredError with stale flag", func(t *testing.T) {
		uc, ur, sr, _ := setup()
		ur.addUser(fixtureUser())
		ur.setMemberships(fixtureUserID, []domain.Membership{m1, m2})

		_, err := uc.Execute(context.Background(), fixtureEmail, fixturePassword, fixtureMembership3.String(), true)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		var selErr *domain.MembershipSelectionRequiredError
		if !isMembershipSelectionRequired(err) {
			t.Fatalf("expected MembershipSelectionRequiredError, got %T: %v", err, err)
		}
		selErr = err.(*domain.MembershipSelectionRequiredError)
		if !selErr.IsStaleChoice {
			t.Fatal("expected IsStaleChoice=true for stale selection")
		}
		if len(selErr.Memberships) != 2 {
			t.Fatalf("expected 2 choices, got %d", len(selErr.Memberships))
		}
		creates := sr.callsByMethod("CreateRefreshToken")
		if len(creates) != 0 {
			t.Fatalf("expected 0 creates, got %d", len(creates))
		}
	})

	t.Run("multi-membership valid selection returns scope claims matching selected membership", func(t *testing.T) {
		uc, ur, _, tk := setup()
		ur.addUser(fixtureUser())
		ur.setMemberships(fixtureUserID, []domain.Membership{m1, m2})

		result, err := uc.Execute(context.Background(), fixtureEmail, fixturePassword, fixtureMembership2.String(), true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.AccessToken != fakeAccessToken {
			t.Fatalf("expected fake access token")
		}
		scope := tk.lastScope()
		if scope.MembershipID != fixtureMembership2.String() {
			t.Fatalf("expected scope membership %s, got %s", fixtureMembership2, scope.MembershipID)
		}
		if scope.TenantID != fixtureTenantID.String() {
			t.Fatalf("expected scope tenant %s, got %s", fixtureTenantID, scope.TenantID)
		}
		if scope.BranchID != fixtureBranchID.String() {
			t.Fatalf("expected scope branch %s, got %s", fixtureBranchID, scope.BranchID)
		}
		if scope.Role != "practitioner" {
			t.Fatalf("expected scope role practitioner, got %s", scope.Role)
		}
	})

	t.Run("result includes all memberships and active membership", func(t *testing.T) {
		uc, ur, _, _ := setup()
		ur.addUser(fixtureUser())
		ur.setMemberships(fixtureUserID, []domain.Membership{m1})

		result, err := uc.Execute(context.Background(), fixtureEmail, fixturePassword, "", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Memberships) != 1 {
			t.Fatalf("expected 1 membership, got %d", len(result.Memberships))
		}
		if result.ActiveMembership.ID != m1.ID {
			t.Fatalf("expected active membership %s, got %s", m1.ID, result.ActiveMembership.ID)
		}
	})

	t.Run("remember_me=true stores RememberMe on refresh token", func(t *testing.T) {
		uc, ur, sr, _ := setup()
		ur.addUser(fixtureUser())
		ur.setMemberships(fixtureUserID, []domain.Membership{m1})

		_, err := uc.Execute(context.Background(), fixtureEmail, fixturePassword, "", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		creates := sr.callsByMethod("CreateRefreshToken")
		if len(creates) != 1 {
			t.Fatalf("expected 1 create, got %d", len(creates))
		}
		if !creates[0].Replacement.RememberMe {
			t.Fatal("expected RememberMe=true on refresh token")
		}
	})

	t.Run("remember_me=false stores RememberMe=false on refresh token", func(t *testing.T) {
		uc, ur, sr, _ := setup()
		ur.addUser(fixtureUser())
		ur.setMemberships(fixtureUserID, []domain.Membership{m1})

		_, err := uc.Execute(context.Background(), fixtureEmail, fixturePassword, "", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		creates := sr.callsByMethod("CreateRefreshToken")
		if len(creates) != 1 {
			t.Fatalf("expected 1 create, got %d", len(creates))
		}
		if creates[0].Replacement.RememberMe {
			t.Fatal("expected RememberMe=false on refresh token")
		}
	})

	t.Run("remember_me defaults to true when not passed", func(t *testing.T) {
		uc, ur, sr, _ := setup()
		ur.addUser(fixtureUser())
		ur.setMemberships(fixtureUserID, []domain.Membership{m1})

		_, err := uc.Execute(context.Background(), fixtureEmail, fixturePassword, "", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		creates := sr.callsByMethod("CreateRefreshToken")
		if len(creates) != 1 {
			t.Fatalf("expected 1 create, got %d", len(creates))
		}
		if !creates[0].Replacement.RememberMe {
			t.Fatal("expected RememberMe=true by default")
		}
	})
}

func assertErrorIs(t *testing.T, got, want error) {
	t.Helper()
	if got == nil {
		t.Fatalf("expected error %v, got nil", want)
	}
	if got.Error() != want.Error() {
		t.Fatalf("expected error %v, got %v", want, got)
	}
}

func isValidationError(err error) bool {
	_, ok := err.(*domain.ValidationError)
	return ok
}
