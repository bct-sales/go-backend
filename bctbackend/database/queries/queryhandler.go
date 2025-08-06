package queries

import (
	"database/sql"
	"errors"
	"fmt"
)

type DatabaseQuerier interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type TransactionalDatabaseQuerier struct {
	transaction *sql.Tx
	committed   bool
}

func NewTransactionDatabaseQuerier(db *sql.DB) (*TransactionalDatabaseQuerier, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start new transaction: %w", err)
	}

	return &TransactionalDatabaseQuerier{
		transaction: tx,
		committed:   false,
	}, nil
}

func (t *TransactionalDatabaseQuerier) Commit() error {
	if err := t.transaction.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	t.committed = true
	return nil
}

func (t *TransactionalDatabaseQuerier) Rollback() error {
	if t.committed {
		return nil
	}

	if err := t.transaction.Rollback(); err != nil {
		return fmt.Errorf("failed to roll back transaction: %w", err)
	}

	return nil
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

	result, err := fn(transaction)
	if err != nil {
		rollbackErr := transaction.Rollback()
		return result, errors.Join(err, rollbackErr)
	} else {
		if err := transaction.Commit(); err != nil {
			return result, err
		}
	}

	return result, nil
}
