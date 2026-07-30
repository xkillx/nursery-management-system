package domain

import (
	"context"

	"github.com/google/uuid"
)

// FakeEnqueuer is a test double that captures enqueued emails.
type FakeEnqueuer struct {
	Enqueued []EnqueueParams
}

func NewFakeEnqueuer() *FakeEnqueuer {
	return &FakeEnqueuer{}
}

func (f *FakeEnqueuer) Enqueue(_ context.Context, _, _ uuid.UUID, params EnqueueParams) (uuid.UUID, error) {
	f.Enqueued = append(f.Enqueued, params)
	return uuid.New(), nil
}
