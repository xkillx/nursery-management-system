package tokens

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateProducesURLSafeNonEmptyToken(t *testing.T) {
	m := NewManager("test-secret", 60)
	tok, err := m.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if tok.Raw == "" {
		t.Fatal("raw token is empty")
	}
	if tok.Hash == "" {
		t.Fatal("hash is empty")
	}
	if strings.ContainsAny(tok.Raw, "+/") {
		t.Fatal("raw token is not URL-safe")
	}
}

func TestSameRawTokenHashesDeterministically(t *testing.T) {
	m := NewManager("test-secret", 60)
	tok, _ := m.Generate()

	hash := m.Hash(tok.Raw)
	if hash != tok.Hash {
		t.Fatalf("expected same hash %q, got %q", tok.Hash, hash)
	}
}

func TestDifferentSecretProducesDifferentHash(t *testing.T) {
	m1 := NewManager("secret-a", 60)
	m2 := NewManager("secret-b", 60)
	tok, _ := m1.Generate()

	h1 := m1.Hash(tok.Raw)
	h2 := m2.Hash(tok.Raw)
	if h1 == h2 {
		t.Fatal("different secrets should produce different hashes")
	}
}

func TestExpiryUsesConfiguredTTL(t *testing.T) {
	fixed := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	m := NewManagerWithClock("test-secret", 60, func() time.Time { return fixed })

	tok, err := m.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	expected := fixed.Add(60 * time.Minute)
	if !tok.ExpiresAt.Equal(expected) {
		t.Fatalf("expected expires_at %v, got %v", expected, tok.ExpiresAt)
	}
}
