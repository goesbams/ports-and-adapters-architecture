package infrastructure

import "context"

// DBTransaction defines a database transaction interface
type DBTransaction interface {
	// BeginTx starts a new transaction and returns a context with the transaction
	BeginTx(ctx context.Context) (context.Context, error)

	// CommitTx commits the transaction in the context
	CommitTx(ctx context.Context) error

	// RollbackTx rolls back the transaction in the context
	RollbackTx(ctx context.Context) error
}
