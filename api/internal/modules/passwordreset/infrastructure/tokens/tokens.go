package tokens

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"
)

type Manager struct {
	secret  []byte
	ttl     time.Duration
	nowFunc func() time.Time
}

func NewManager(secret string, ttlMinutes int) *Manager {
	return &Manager{
		secret:  []byte(secret),
		ttl:     time.Duration(ttlMinutes) * time.Minute,
		nowFunc: func() time.Time { return time.Now().UTC() },
	}
}

func NewManagerWithClock(secret string, ttlMinutes int, nowFunc func() time.Time) *Manager {
	return &Manager{
		secret:  []byte(secret),
		ttl:     time.Duration(ttlMinutes) * time.Minute,
		nowFunc: nowFunc,
	}
}

type Token struct {
	Raw       string
	Hash      string
	ExpiresAt time.Time
}

func (m *Manager) Generate() (Token, error) {
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return Token{}, err
	}

	raw := base64.RawURLEncoding.EncodeToString(rawBytes)
	hash := m.Hash(raw)
	expiresAt := m.nowFunc().Add(m.ttl)

	return Token{
		Raw:       raw,
		Hash:      hash,
		ExpiresAt: expiresAt,
	}, nil
}

func (m *Manager) Hash(raw string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(raw))
	return hex.EncodeToString(mac.Sum(nil))
}
