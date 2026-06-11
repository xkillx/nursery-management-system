package application

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	MinPasswordLen = 4
	MaxPasswordLen = 64
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
