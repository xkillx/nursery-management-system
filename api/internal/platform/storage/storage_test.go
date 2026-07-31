package storage

import (
	"context"
	"fmt"
	"testing"
)

func TestFakeService_UploadDownload(t *testing.T) {
	svc := NewFakeService()
	ctx := context.Background()

	err := svc.Upload(ctx, "test-key", []byte("hello world"), "text/plain")
	if err != nil {
		t.Fatalf("upload: %v", err)
	}

	data, err := svc.Download(ctx, "test-key")
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("got %q, want %q", string(data), "hello world")
	}
}

func TestFakeService_DownloadNotFound(t *testing.T) {
	svc := NewFakeService()
	_, err := svc.Download(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestFakeService_Delete(t *testing.T) {
	svc := NewFakeService()
	ctx := context.Background()

	svc.Upload(ctx, "key1", []byte("data"), "application/octet-stream")
	if err := svc.Delete(ctx, "key1"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err := svc.Download(ctx, "key1")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestFakeService_UploadError(t *testing.T) {
	svc := NewFakeService()
	svc.UploadErr = fmt.Errorf("upload failed")
	err := svc.Upload(context.Background(), "key", []byte("data"), "text/plain")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFakeService_DownloadError(t *testing.T) {
	svc := NewFakeService()
	svc.DownloadErr = fmt.Errorf("download failed")
	_, err := svc.Download(context.Background(), "key")
	if err == nil {
		t.Fatal("expected error")
	}
}
