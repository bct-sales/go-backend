package queries

import (
	"database/sql"
	"fmt"
)

type DatabaseQuerier interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type TransactionedDatabaseQuerier struct {
	transaction *sql.Tx
	committed   bool
}

func NewTransactionDatabaseQuerier(db *sql.DB) (*TransactionedDatabaseQuerier, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start new transaction: %w", err)
	}

	return &TransactionedDatabaseQuerier{
		transaction: tx,
		committed:   false,
	}, nil
}

func (t *TransactionedDatabaseQuerier) Commit() error {
	if err := t.transaction.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	t.committed = true
	return nil
}

func (t *TransactionedDatabaseQuerier) Rollback() error {
	if t.committed {
		return nil
	}

	if err := t.transaction.Rollback(); err != nil {
		return fmt.Errorf("failed to roll back transaction: %w", err)
	}

	return nil
}

func (t *TransactionedDatabaseQuerier) Exec(query string, args ...any) (sql.Result, error) {
	return t.transaction.Exec(query, args...)
}

func (t *TransactionedDatabaseQuerier) Query(query string, args ...any) (*sql.Rows, error) {
	return t.transaction.Query(query, args...)
}

func (t *TransactionedDatabaseQuerier) QueryRow(query string, args ...any) *sql.Row {
	return t.transaction.QueryRow(query, args...)
}
