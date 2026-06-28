package events_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nursery-management-system/api/internal/platform/events"
)

type testEvent struct {
	Value string
	At    time.Time
}

func (e testEvent) OccurredAt() time.Time { return e.At }

type testEvent2 struct {
	Count int
	At    time.Time
}

func (e testEvent2) OccurredAt() time.Time { return e.At }

type mockTxManager struct {
	fn func(ctx context.Context, txFn func(tx pgx.Tx) error) error
}

func (m *mockTxManager) ExecTx(ctx context.Context, txFn func(tx pgx.Tx) error) error {
	return m.fn(ctx, txFn)
}

func TestEventDispatcher_EmitsToRegisteredHandler(t *testing.T) {
	var handled atomic.Bool
	mgr := &mockTxManager{
		fn: func(ctx context.Context, txFn func(tx pgx.Tx) error) error {
			return txFn(nil)
		},
	}

	dispatcher := events.NewEventDispatcher(mgr)
	events.Register(dispatcher, events.TypedHandlerFunc[testEvent](func(ctx context.Context, tx pgx.Tx, event testEvent) error {
		handled.Store(true)
		assert.Equal(t, "hello", event.Value)
		return nil
	}))

	err := dispatcher.DispatchInTx(context.Background(), func(emitter events.Emitter) error {
		emitter.Emit(testEvent{Value: "hello", At: time.Now()})
		return nil
	})

	require.NoError(t, err)
	assert.True(t, handled.Load())
}

func TestEventDispatcher_NoRegisteredHandlerIsNoOp(t *testing.T) {
	mgr := &mockTxManager{
		fn: func(ctx context.Context, txFn func(tx pgx.Tx) error) error {
			return txFn(nil)
		},
	}

	dispatcher := events.NewEventDispatcher(mgr)

	err := dispatcher.DispatchInTx(context.Background(), func(emitter events.Emitter) error {
		emitter.Emit(testEvent{Value: "no handler", At: time.Now()})
		return nil
	})

	require.NoError(t, err)
}

func TestEventDispatcher_HandlerFailureReturnsError(t *testing.T) {
	expectedErr := errors.New("handler failed")
	mgr := &mockTxManager{
		fn: func(ctx context.Context, txFn func(tx pgx.Tx) error) error {
			return txFn(nil)
		},
	}

	dispatcher := events.NewEventDispatcher(mgr)
	events.Register(dispatcher, events.TypedHandlerFunc[testEvent](func(ctx context.Context, tx pgx.Tx, event testEvent) error {
		return expectedErr
	}))

	err := dispatcher.DispatchInTx(context.Background(), func(emitter events.Emitter) error {
		emitter.Emit(testEvent{Value: "fail", At: time.Now()})
		return nil
	})

	require.Error(t, err)
}

func TestEventDispatcher_MultipleHandlersInOrder(t *testing.T) {
	var callOrder []int
	mgr := &mockTxManager{
		fn: func(ctx context.Context, txFn func(tx pgx.Tx) error) error {
			return txFn(nil)
		},
	}

	dispatcher := events.NewEventDispatcher(mgr)
	events.Register(dispatcher, events.TypedHandlerFunc[testEvent](func(ctx context.Context, tx pgx.Tx, event testEvent) error {
		callOrder = append(callOrder, 1)
		return nil
	}))
	events.Register(dispatcher, events.TypedHandlerFunc[testEvent](func(ctx context.Context, tx pgx.Tx, event testEvent) error {
		callOrder = append(callOrder, 2)
		return nil
	}))

	err := dispatcher.DispatchInTx(context.Background(), func(emitter events.Emitter) error {
		emitter.Emit(testEvent{Value: "multi", At: time.Now()})
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, []int{1, 2}, callOrder)
}

func TestEventDispatcher_MainWorkFailureSkipsHandlers(t *testing.T) {
	var handled atomic.Bool
	mgr := &mockTxManager{
		fn: func(ctx context.Context, txFn func(tx pgx.Tx) error) error {
			return txFn(nil)
		},
	}

	dispatcher := events.NewEventDispatcher(mgr)
	events.Register(dispatcher, events.TypedHandlerFunc[testEvent](func(ctx context.Context, tx pgx.Tx, event testEvent) error {
		handled.Store(true)
		return nil
	}))

	err := dispatcher.DispatchInTx(context.Background(), func(emitter events.Emitter) error {
		emitter.Emit(testEvent{Value: "fail", At: time.Now()})
		return errors.New("main work failed")
	})

	require.Error(t, err)
	assert.False(t, handled.Load())
}

func TestEventDispatcher_DifferentEventTypes(t *testing.T) {
	var handled1, handled2 atomic.Bool
	mgr := &mockTxManager{
		fn: func(ctx context.Context, txFn func(tx pgx.Tx) error) error {
			return txFn(nil)
		},
	}

	dispatcher := events.NewEventDispatcher(mgr)
	events.Register(dispatcher, events.TypedHandlerFunc[testEvent](func(ctx context.Context, tx pgx.Tx, event testEvent) error {
		handled1.Store(true)
		return nil
	}))
	events.Register(dispatcher, events.TypedHandlerFunc[testEvent2](func(ctx context.Context, tx pgx.Tx, event testEvent2) error {
		handled2.Store(true)
		return nil
	}))

	err := dispatcher.DispatchInTx(context.Background(), func(emitter events.Emitter) error {
		emitter.Emit(testEvent{Value: "type1", At: time.Now()})
		emitter.Emit(testEvent2{Count: 42, At: time.Now()})
		return nil
	})

	require.NoError(t, err)
	assert.True(t, handled1.Load())
	assert.True(t, handled2.Load())
}

func TestTypedHandlerFunc_Convenience(t *testing.T) {
	fn := events.TypedHandlerFunc[testEvent](func(ctx context.Context, tx pgx.Tx, event testEvent) error {
		return nil
	})
	var _ events.TypedHandler[testEvent] = fn
	assert.NotNil(t, fn)
}

func testDBPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), "postgres://localhost:5432/nursery_test?sslmode=disable")
	if err != nil {
		t.Skipf("skipping integration test: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestEventDispatcher_IntegrationTransactionalRollback(t *testing.T) {
	pool := testDBPool(t)

	mgr := &mockTxManager{
		fn: func(ctx context.Context, txFn func(tx pgx.Tx) error) error {
			tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				return fmt.Errorf("begin tx: %w", err)
			}
			if err := txFn(tx); err != nil {
				_ = tx.Rollback(ctx)
				return err
			}
			return tx.Commit(ctx)
		},
	}

	dispatcher := events.NewEventDispatcher(mgr)
	events.Register(dispatcher, events.TypedHandlerFunc[testEvent](func(ctx context.Context, tx pgx.Tx, event testEvent) error {
		return errors.New("handler rollback")
	}))

	ctx := context.Background()

	err := dispatcher.DispatchInTx(ctx, func(emitter events.Emitter) error {
		emitter.Emit(testEvent{Value: "integration", At: time.Now()})
		return nil
	})

	require.Error(t, err)
}
