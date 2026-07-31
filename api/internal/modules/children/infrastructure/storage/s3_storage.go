package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/children/domain"
	"nursery-management-system/api/internal/platform/storage"
)

const (
	s3KeyPrefix        = "children/"
	presignedURLExpiry = 1 * time.Hour
)

// S3FileStorage implements domain.FileStorage by storing child photos in S3
// via the platform storage.Service abstraction.
type S3FileStorage struct {
	service storage.Service
}

// NewS3FileStorage creates a new S3FileStorage backed by the given storage.Service.
func NewS3FileStorage(service storage.Service) *S3FileStorage {
	return &S3FileStorage{service: service}
}

// Save reads the entire file from data, uploads it to S3 with the key structure
// children/<tenantID>/<branchID>/<childID>/photo.<ext>, and returns the key.
func (s *S3FileStorage) Save(ctx context.Context, tenantID, branchID, childID uuid.UUID, data io.Reader, ext string) (string, error) {
	buf, err := io.ReadAll(io.LimitReader(data, maxFileSize+1))
	if err != nil {
		return "", fmt.Errorf("read file data: %w", err)
	}
	if int64(len(buf)) > maxFileSize {
		return "", fmt.Errorf("file exceeds maximum size of %d bytes", maxFileSize)
	}

	key := s3KeyPrefix + tenantID.String() + "/" + branchID.String() + "/" + childID.String() + "/photo." + ext

	contentType := "application/octet-stream"
	switch ext {
	case "jpg", "jpeg":
		contentType = "image/jpeg"
	case "png":
		contentType = "image/png"
	}

	if err := s.service.Upload(ctx, key, buf, contentType); err != nil {
		return "", fmt.Errorf("upload to s3: %w", err)
	}

	return key, nil
}

// Delete removes the S3 object at the given key.
func (s *S3FileStorage) Delete(ctx context.Context, path string) error {
	if path == "" {
		return nil
	}
	if err := s.service.Delete(ctx, path); err != nil {
		return fmt.Errorf("delete from s3: %w", err)
	}
	return nil
}

// GetURL returns a presigned S3 URL for the given key with 1 hour expiry.
func (s *S3FileStorage) GetURL(ctx context.Context, path string) (string, error) {
	if path == "" {
		return "", nil
	}
	url, err := s.service.GetPresignedURL(ctx, path, presignedURLExpiry)
	if err != nil {
		return "", fmt.Errorf("get presigned url: %w", err)
	}
	return url, nil
}

// Ensure S3FileStorage implements domain.FileStorage.
var _ domain.FileStorage = (*S3FileStorage)(nil)
