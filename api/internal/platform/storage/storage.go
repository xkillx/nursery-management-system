package storage

import (
	"context"
	"time"
)

type Service interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) error
	Download(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}
