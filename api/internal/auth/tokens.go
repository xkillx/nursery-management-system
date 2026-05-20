package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenManager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

type AccessClaims struct {
	jwt.RegisteredClaims
	Email        string `json:"email"`
	MembershipID string `json:"membership_id"`
	TenantID     string `json:"tenant_id"`
	BranchID     string `json:"branch_id"`
	Role         string `json:"role"`
}

type ScopeClaims struct {
	MembershipID string
	TenantID     string
	BranchID     string
	Role         string
}

func NewTokenManager(accessSecret, refreshSecret string, accessTTLMin, refreshTTLHours int) *TokenManager {
	return &TokenManager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     time.Duration(accessTTLMin) * time.Minute,
		refreshTTL:    time.Duration(refreshTTLHours) * time.Hour,
	}
}

func (m *TokenManager) NewAccessToken(userID uuid.UUID, email string, scope ScopeClaims) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(m.accessTTL)

	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		Email:        email,
		MembershipID: scope.MembershipID,
		TenantID:     scope.TenantID,
		BranchID:     scope.BranchID,
		Role:         scope.Role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.accessSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}

	return signed, expiresAt, nil
}

func (m *TokenManager) ParseAccessToken(raw string) (AccessClaims, error) {
	claims := AccessClaims{}
	token, err := jwt.ParseWithClaims(raw, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidAccessToken
		}
		return m.accessSecret, nil
	})
	if err != nil {
		return AccessClaims{}, ErrInvalidAccessToken
	}
	if !token.Valid {
		return AccessClaims{}, ErrInvalidAccessToken
	}

	if claims.Subject == "" ||
		claims.MembershipID == "" ||
		claims.TenantID == "" ||
		claims.BranchID == "" ||
		claims.Role == "" ||
		claims.ExpiresAt == nil ||
		claims.IssuedAt == nil {
		return AccessClaims{}, ErrInvalidAccessToken
	}

	return claims, nil
}

func (m *TokenManager) NewRefreshToken() (raw string, hash string, expiresAt time.Time, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", time.Time{}, fmt.Errorf("generate refresh token: %w", err)
	}

	raw = base64.RawURLEncoding.EncodeToString(b)
	hash = hashRefreshToken(raw, m.refreshSecret)
	expiresAt = time.Now().UTC().Add(m.refreshTTL)
	return raw, hash, expiresAt, nil
}

func (m *TokenManager) HashRefreshToken(raw string) string {
	return hashRefreshToken(raw, m.refreshSecret)
}

func (m *TokenManager) AccessTTLSeconds() int64 {
	return int64(m.accessTTL.Seconds())
}

func hashRefreshToken(raw string, secret []byte) string {
	sum := sha256.Sum256(append([]byte(raw), secret...))
	return hex.EncodeToString(sum[:])
}
