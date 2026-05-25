package tokens

import (
	"testing"
	"time"
)

func TestGenerateProducesToken(t *testing.T) {
	m := NewManager("test-secret", 168)
	tok, err := m.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if tok.Raw == "" {
		t.Fatal("Raw token is empty")
	}
	if tok.Hash == "" {
		t.Fatal("Hash is empty")
	}
	if tok.ExpiresAt.IsZero() {
		t.Fatal("ExpiresAt is zero")
	}
}

func TestHashIsDeterministic(t *testing.T) {
	m := NewManager("test-secret", 168)
	tok, _ := m.Generate()
	h1 := m.Hash(tok.Raw)
	h2 := m.Hash(tok.Raw)
	if h1 != h2 {
		t.Fatal("Hash is not deterministic")
	}
}

func TestGenerateMatchesHash(t *testing.T) {
	m := NewManager("test-secret", 168)
	tok, _ := m.Generate()
	if tok.Hash != m.Hash(tok.Raw) {
		t.Fatal("Token hash does not match Hash(raw)")
	}
}

func TestDifferentSecretsProduceDifferentHashes(t *testing.T) {
	m1 := NewManager("secret-a", 168)
	m2 := NewManager("secret-b", 168)
	tok, _ := m1.Generate()
	if m1.Hash(tok.Raw) == m2.Hash(tok.Raw) {
		t.Fatal("Different secrets produced same hash")
	}
}

func TestExpiryUsesTTLCorrectly(t *testing.T) {
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	m := NewManagerWithClock("test-secret", 168, func() time.Time { return fixedTime })
	tok, _ := m.Generate()
	expected := fixedTime.Add(168 * time.Hour)
	if !tok.ExpiresAt.Equal(expected) {
		t.Fatalf("ExpiresAt = %v, want %v", tok.ExpiresAt, expected)
	}
}

func TestEachGenerateProducesUniqueRaw(t *testing.T) {
	m := NewManager("test-secret", 168)
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		tok, _ := m.Generate()
		if seen[tok.Raw] {
			t.Fatal("duplicate raw token generated")
		}
		seen[tok.Raw] = true
	}
}
