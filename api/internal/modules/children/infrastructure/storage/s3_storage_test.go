package storage

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"

	"nursery-management-system/api/internal/platform/storage"
)

func TestS3FileStorage_Save(t *testing.T) {
	fake := storage.NewFakeService()
	s := NewS3FileStorage(fake)

	tenantID := uuid.New()
	branchID := uuid.New()
	childID := uuid.New()

	data := bytes.NewReader([]byte("fake-photo-data"))
	key, err := s.Save(context.Background(), tenantID, branchID, childID, data, "jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedKey := "children/" + tenantID.String() + "/" + branchID.String() + "/" + childID.String() + "/photo.jpg"
	if key != expectedKey {
		t.Errorf("got key %q, want %q", key, expectedKey)
	}

	stored, ok := fake.Objects[key]
	if !ok {
		t.Fatalf("object not found in fake storage at key %q", key)
	}
	if string(stored) != "fake-photo-data" {
		t.Errorf("got data %q, want %q", string(stored), "fake-photo-data")
	}
}

func TestS3FileStorage_Save_ExceedsSizeLimit(t *testing.T) {
	fake := storage.NewFakeService()
	s := NewS3FileStorage(fake)

	bigData := strings.Repeat("x", maxFileSize+1)
	_, err := s.Save(context.Background(), uuid.New(), uuid.New(), uuid.New(), strings.NewReader(bigData), "jpg")
	if err == nil {
		t.Fatal("expected error for oversized file")
	}
	if !strings.Contains(err.Error(), "exceeds maximum size") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestS3FileStorage_Save_ReadError(t *testing.T) {
	fake := storage.NewFakeService()
	s := NewS3FileStorage(fake)

	errReader := &errReader{err: io.ErrUnexpectedEOF}
	_, err := s.Save(context.Background(), uuid.New(), uuid.New(), uuid.New(), errReader, "jpg")
	if err == nil {
		t.Fatal("expected error for read failure")
	}
}

func TestS3FileStorage_Delete(t *testing.T) {
	fake := storage.NewFakeService()
	fake.Objects["children/tenant/branch/child/photo.jpg"] = []byte("data")
	s := NewS3FileStorage(fake)

	err := s.Delete(context.Background(), "children/tenant/branch/child/photo.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := fake.Objects["children/tenant/branch/child/photo.jpg"]; ok {
		t.Error("expected object to be deleted")
	}
}

func TestS3FileStorage_Delete_EmptyPath(t *testing.T) {
	fake := storage.NewFakeService()
	s := NewS3FileStorage(fake)

	err := s.Delete(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestS3FileStorage_GetURL(t *testing.T) {
	fake := storage.NewFakeService()
	s := NewS3FileStorage(fake)

	url, err := s.GetURL(context.Background(), "children/tenant/branch/child/photo.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Error("expected non-empty URL")
	}
	if !strings.Contains(url, "children/tenant/branch/child/photo.jpg") {
		t.Errorf("URL %q does not contain key", url)
	}
}

func TestS3FileStorage_GetURL_EmptyPath(t *testing.T) {
	fake := storage.NewFakeService()
	s := NewS3FileStorage(fake)

	url, err := s.GetURL(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "" {
		t.Errorf("expected empty URL, got %q", url)
	}
}

type errReader struct {
	err error
}

func (r *errReader) Read([]byte) (int, error) {
	return 0, r.err
}
