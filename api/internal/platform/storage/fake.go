package storage

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type FakeService struct {
	mu              sync.Mutex
	Objects         map[string][]byte
	UploadErr       error
	DownloadErr     error
	DeleteErr       error
	PresignedURLErr error
}

func NewFakeService() *FakeService {
	return &FakeService{
		Objects: make(map[string][]byte),
	}
}

func (f *FakeService) Upload(_ context.Context, key string, data []byte, _ string) error {
	if f.UploadErr != nil {
		return f.UploadErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Objects[key] = data
	return nil
}

func (f *FakeService) Download(_ context.Context, key string) ([]byte, error) {
	if f.DownloadErr != nil {
		return nil, f.DownloadErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	data, ok := f.Objects[key]
	if !ok {
		return nil, fmt.Errorf("object not found: %s", key)
	}
	return data, nil
}

func (f *FakeService) Delete(_ context.Context, key string) error {
	if f.DeleteErr != nil {
		return f.DeleteErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.Objects, key)
	return nil
}

func (f *FakeService) GetPresignedURL(_ context.Context, key string, _ time.Duration) (string, error) {
	if f.PresignedURLErr != nil {
		return "", f.PresignedURLErr
	}
	return "https://fake-s3.example.com/" + key, nil
}
