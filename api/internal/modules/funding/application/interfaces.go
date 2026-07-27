package application

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// TxManager is the interface satisfied by *transaction.Manager. The
// application layer depends on this interface so it can be mocked in tests.
type TxManager interface {
	ExecTx(ctx context.Context, fn func(tx pgx.Tx) error) error
}
