package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/modules/children/domain"
)

const (
	maxFileSize = 5 * 1024 * 1024 // 5 MB
	uploadDir   = "uploads/child-photos"
)

// LocalStorage implements domain.FileStorage by writing files to local disk.
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new LocalStorage rooted at the given base path.
func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{basePath: basePath}
}

// Save writes the file to disk at uploads/child-photos/<tenant>/<branch>/<child>/photo.<ext>.
// It enforces the 5 MB size limit.
func (s *LocalStorage) Save(ctx context.Context, tenantID, branchID, childID uuid.UUID, data io.Reader, ext string) (string, error) {
	dir := filepath.Join(s.basePath, uploadDir, tenantID.String(), branchID.String(), childID.String())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create upload directory: %w", err)
	}

	filename := "photo." + ext
	fullPath := filepath.Join(dir, filename)

	f, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	written, err := io.Copy(f, io.LimitReader(data, maxFileSize+1))
	if err != nil {
		os.Remove(fullPath)
		return "", fmt.Errorf("write file: %w", err)
	}
	if written > maxFileSize {
		os.Remove(fullPath)
		return "", fmt.Errorf("file exceeds maximum size of %d bytes", maxFileSize)
	}

	relPath := filepath.Join(uploadDir, tenantID.String(), branchID.String(), childID.String(), filename)
	return relPath, nil
}

// Delete removes the file at the given relative path.
func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	if path == "" {
		return nil
	}
	fullPath := filepath.Join(s.basePath, path)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

// GetURL returns the relative API path for serving the file.
func (s *LocalStorage) GetURL(ctx context.Context, path string) (string, error) {
	if path == "" {
		return "", nil
	}
	return "/api/v1/children/photo?path=" + path, nil
}

// Ensure LocalStorage implements domain.FileStorage.
var _ domain.FileStorage = (*LocalStorage)(nil)
