package events

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/jackc/pgx/v5"
)

type DomainEvent interface {
	OccurredAt() time.Time
}

type Handler interface {
	Handle(ctx context.Context, tx pgx.Tx, event DomainEvent) error
}

type TypedHandler[T DomainEvent] interface {
	Handle(ctx context.Context, tx pgx.Tx, event T) error
}

type TypedHandlerFunc[T DomainEvent] func(ctx context.Context, tx pgx.Tx, event T) error

func (f TypedHandlerFunc[T]) Handle(ctx context.Context, tx pgx.Tx, event T) error {
	return f(ctx, tx, event)
}

type typedAdapter[T DomainEvent] struct {
	inner TypedHandler[T]
}

func (a *typedAdapter[T]) Handle(ctx context.Context, tx pgx.Tx, event DomainEvent) error {
	return a.inner.Handle(ctx, tx, event.(T))
}

func AsHandler[T DomainEvent](handler TypedHandler[T]) Handler {
	return &typedAdapter[T]{inner: handler}
}

type Emitter interface {
	Emit(event DomainEvent)
}

type emitter struct {
	events []DomainEvent
}

func (e *emitter) Emit(event DomainEvent) {
	e.events = append(e.events, event)
}

type txManager interface {
	ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}

type handlerEntry struct {
	eventType reflect.Type
	handler   Handler
}

type EventDispatcher struct {
	txMgr   txManager
	entries []handlerEntry
}

func NewEventDispatcher(txMgr interface {
	ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}) *EventDispatcher {
	return &EventDispatcher{txMgr: txMgr, entries: make([]handlerEntry, 0)}
}

func (d *EventDispatcher) DispatchInTx(ctx context.Context, fn func(tx pgx.Tx, emitter Emitter) error) error {
	em := &emitter{events: make([]DomainEvent, 0)}

	err := d.txMgr.ExecTx(ctx, func(tx pgx.Tx) error {
		if err := fn(tx, em); err != nil {
			return err
		}

		for _, event := range em.events {
			eventType := reflect.TypeOf(event)
			for _, entry := range d.entries {
				if entry.eventType == eventType {
					if err := entry.handler.Handle(ctx, tx, event); err != nil {
						return fmt.Errorf("event handler for %T: %w", event, err)
					}
				}
			}
		}
		return nil
	})

	return err
}

func Register[T DomainEvent](d *EventDispatcher, handler TypedHandler[T]) {
	var zero T
	d.entries = append(d.entries, handlerEntry{
		eventType: reflect.TypeOf(zero),
		handler:   AsHandler(handler),
	})
}
