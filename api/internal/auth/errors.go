package auth

import "errors"

var ErrNotFound = errors.New("not found")

var ErrInvalidAccessToken = errors.New("invalid access token")
