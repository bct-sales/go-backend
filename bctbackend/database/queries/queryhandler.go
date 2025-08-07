package queries

import (
	"database/sql"
	"fmt"
	"log/slog"
)

type DatabaseQuerier interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type TransactionalDatabaseQuerier struct {
	transaction *sql.Tx
}

func NewTransactionDatabaseQuerier(db *sql.DB) (*TransactionalDatabaseQuerier, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start new transaction: %w", err)
	}

	querier := TransactionalDatabaseQuerier{transaction: tx}
	return &querier, nil
}

func (t *TransactionalDatabaseQuerier) Commit() error {
	if err := t.transaction.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (t *TransactionalDatabaseQuerier) Rollback() {
	if err := t.transaction.Rollback(); err != nil {
		slog.Error("failed to roll back transaction", "error", err)
	}
}

func (t *TransactionalDatabaseQuerier) Exec(query string, args ...any) (sql.Result, error) {
	return t.transaction.Exec(query, args...)
}

func (t *TransactionalDatabaseQuerier) Query(query string, args ...any) (*sql.Rows, error) {
	return t.transaction.Query(query, args...)
}

func (t *TransactionalDatabaseQuerier) QueryRow(query string, args ...any) *sql.Row {
	return t.transaction.QueryRow(query, args...)
}

func WithTransaction[T any](db *sql.DB, fn func(transaction *TransactionalDatabaseQuerier) (T, error)) (T, error) {
	transaction, err := NewTransactionDatabaseQuerier(db)
	if err != nil {
		var dummy T
		return dummy, err
	}
	defer transaction.Rollback()

	result, err := fn(transaction)
	if err != nil {
		return result, err
	}

	if err := transaction.Commit(); err != nil {
		return result, err
	}

	return result, nil
}
