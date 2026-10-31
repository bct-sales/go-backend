package queries

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/Masterminds/squirrel"
)

type DatabaseQuerier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// squirrelRunner adapts a DatabaseQuerier to squirrel's RunWith, which requires
// the legacy context-less Exec and Query methods. The ctx-less methods forward to
// the querier's ctx methods with the context given here; prefer calling the
// Context-suffixed execution methods (e.g. .ExecContext(ctx)) on the runner.
type squirrelRunner struct {
	ctx     context.Context
	querier DatabaseQuerier
}

func newSquirrelRunner(ctx context.Context, querier DatabaseQuerier) squirrelRunner {
	return squirrelRunner{ctx: ctx, querier: querier}
}

func (r squirrelRunner) Exec(query string, args ...any) (sql.Result, error) {
	return r.querier.ExecContext(r.ctx, query, args...)
}

func (r squirrelRunner) Query(query string, args ...any) (*sql.Rows, error) {
	return r.querier.QueryContext(r.ctx, query, args...)
}

func (r squirrelRunner) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return r.querier.ExecContext(ctx, query, args...)
}

func (r squirrelRunner) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return r.querier.QueryContext(ctx, query, args...)
}

func (r squirrelRunner) QueryRowContext(ctx context.Context, query string, args ...any) squirrel.RowScanner {
	return r.querier.QueryRowContext(ctx, query, args...)
}

type TransactionalDatabaseQuerier struct {
	transaction *sql.Tx
	committed   bool
}

func NewTransactionalDatabaseQuerier(ctx context.Context, db *sql.DB) (*TransactionalDatabaseQuerier, error) {
	transaction, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start new transaction: %w", err)
	}

	querier := TransactionalDatabaseQuerier{
		transaction: transaction,
		committed:   false,
	}
	return &querier, nil
}

func (t *TransactionalDatabaseQuerier) Commit() error {
	if t.committed {
		slog.Warn("Commit called on already committed transaction")
	}

	if err := t.transaction.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	t.committed = true
	return nil
}

func (t *TransactionalDatabaseQuerier) RollbackIfNotCommitted() {
	if !t.committed {
		err := t.transaction.Rollback()
		if err != nil {
			slog.Error("Failed to rollback transaction", "error", err)
		}
	}
}

func (t *TransactionalDatabaseQuerier) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	if t.committed {
		slog.Warn("PrepareContext called on already committed transaction")
	}

	return t.transaction.PrepareContext(ctx, query)
}

func (t *TransactionalDatabaseQuerier) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if t.committed {
		slog.Warn("ExecContext called on already committed transaction")
	}

	return t.transaction.ExecContext(ctx, query, args...)
}

func (t *TransactionalDatabaseQuerier) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if t.committed {
		slog.Warn("QueryContext called on already committed transaction")
	}

	return t.transaction.QueryContext(ctx, query, args...)
}

func (t *TransactionalDatabaseQuerier) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if t.committed {
		slog.Warn("QueryRowContext called on already committed transaction")
	}

	return t.transaction.QueryRowContext(ctx, query, args...)
}

func WithTransaction[T any](ctx context.Context, db *sql.DB, fn func(transaction *TransactionalDatabaseQuerier) (T, error)) (T, error) {
	transaction, err := NewTransactionalDatabaseQuerier(ctx, db)
	if err != nil {
		var dummy T
		return dummy, err
	}
	defer transaction.RollbackIfNotCommitted()

	result, err := fn(transaction)
	if err != nil {
		return result, err
	}

	if err := transaction.Commit(); err != nil {
		return result, err
	}

	return result, nil
}
