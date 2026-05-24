package application

import (
	"context"
	"testing"
)

func TestLogoutUseCase_Execute(t *testing.T) {
	setup := func() (*LogoutUseCase, *fakeSessionRepo, *fakeTokens) {
		sr := newFakeSessionRepo()
		tk := newFakeTokens()
		uc := NewLogoutUseCase(sr, tk)
		return uc, sr, tk
	}

	t.Run("hashes raw token and calls RevokeByTokenHash", func(t *testing.T) {
		uc, sr, _ := setup()
		err := uc.Execute(context.Background(), "raw-refresh-token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		revokes := sr.callsByMethod("RevokeByTokenHash")
		if len(revokes) != 1 {
			t.Fatalf("expected 1 revoke, got %d", len(revokes))
		}
		if revokes[0].TokenHash != "hash:raw-refresh-token" {
			t.Fatalf("expected hash:raw-refresh-token, got %q", revokes[0].TokenHash)
		}
	})

	t.Run("returns nil even when repository revoke fails", func(t *testing.T) {
		uc, sr, _ := setup()
		sr.revokeErr = someError()
		err := uc.Execute(context.Background(), "raw-refresh-token")
		if err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("calling twice with same token records two revoke attempts", func(t *testing.T) {
		uc, sr, _ := setup()
		_ = uc.Execute(context.Background(), "raw-refresh-token")
		_ = uc.Execute(context.Background(), "raw-refresh-token")
		revokes := sr.callsByMethod("RevokeByTokenHash")
		if len(revokes) != 2 {
			t.Fatalf("expected 2 revoke calls, got %d", len(revokes))
		}
	})
}

func someError() error {
	return errSome
}

var errSome = func() error {
	return &testError{"some error"}
}()

type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }
