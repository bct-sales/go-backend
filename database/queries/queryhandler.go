package queries

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
)

type DatabaseQuerier interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type ContextDatabaseQuerier struct {
	Database *sql.DB
	Context  context.Context
}

func NewContextDatabaseQuerier(database *sql.DB, ctx context.Context) *ContextDatabaseQuerier {
	return &ContextDatabaseQuerier{
		Database: database,
		Context:  ctx,
	}
}

func (c *ContextDatabaseQuerier) Exec(query string, args ...any) (sql.Result, error) {
	return c.Database.ExecContext(c.Context, query, args...)
}

func (c *ContextDatabaseQuerier) Query(query string, args ...any) (*sql.Rows, error) {
	return c.Database.QueryContext(c.Context, query, args...)
}

func (c *ContextDatabaseQuerier) QueryRow(query string, args ...any) *sql.Row {
	return c.Database.QueryRowContext(c.Context, query, args...)
}

type TransactionalDatabaseQuerier struct {
	transaction *sql.Tx
	context     context.Context
	committed   bool
}

func NewTransactionalDatabaseQuerier(context context.Context, db *sql.DB) (*TransactionalDatabaseQuerier, error) {
	transaction, err := db.BeginTx(context, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to start new transaction: %w", err)
	}

	querier := TransactionalDatabaseQuerier{
		transaction: transaction,
		context:     context,
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

func (t *TransactionalDatabaseQuerier) Exec(query string, args ...any) (sql.Result, error) {
	if t.committed {
		slog.Warn("Exec called on already committed transaction")
	}

	return t.transaction.ExecContext(t.context, query, args...)
}

func (t *TransactionalDatabaseQuerier) Query(query string, args ...any) (*sql.Rows, error) {
	if t.committed {
		slog.Warn("Query called on already committed transaction")
	}

	return t.transaction.QueryContext(t.context, query, args...)
}

func (t *TransactionalDatabaseQuerier) QueryRow(query string, args ...any) *sql.Row {
	if t.committed {
		slog.Warn("QueryRow called on already committed transaction")
	}

	return t.transaction.QueryRowContext(t.context, query, args...)
}

func WithTransaction[T any](context context.Context, db *sql.DB, fn func(transaction *TransactionalDatabaseQuerier) (T, error)) (T, error) {
	transaction, err := NewTransactionalDatabaseQuerier(context, db)
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
