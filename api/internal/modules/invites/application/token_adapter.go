package application

import (
	"time"

	"nursery-management-system/api/internal/modules/invites/infrastructure/tokens"
)

type TokenGeneratorAdapter struct {
	manager *tokens.Manager
}

func NewTokenGeneratorAdapter(manager *tokens.Manager) *TokenGeneratorAdapter {
	return &TokenGeneratorAdapter{manager: manager}
}

func (a *TokenGeneratorAdapter) Generate() (string, string, time.Time, error) {
	tok, err := a.manager.Generate()
	if err != nil {
		return "", "", time.Time{}, err
	}
	return tok.Raw, tok.Hash, tok.ExpiresAt, nil
}
