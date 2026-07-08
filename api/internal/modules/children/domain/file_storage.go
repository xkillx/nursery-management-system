package domain

import (
	"context"
	"io"

	"github.com/google/uuid"
)

// FileStorage abstracts file operations for child profile photos.
// The initial implementation writes to local disk; a future S3
// implementation can replace it without changing the API contract.
type FileStorage interface {
	// Save stores the file content and returns the relative storage path.
	Save(ctx context.Context, tenantID, branchID, childID uuid.UUID, data io.Reader, ext string) (string, error)

	// Delete removes the file at the given path.
	Delete(ctx context.Context, path string) error

	// GetURL returns a servable URL for the given path.
	// For local storage this is a relative API path; for S3 it would be a signed URL.
	GetURL(ctx context.Context, path string) (string, error)
}
