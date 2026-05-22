package uid

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/google/uuid"
)

func NewCSRFToken() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return uuid.NewString()
	}
	return base64.RawURLEncoding.EncodeToString(buf)
}
