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

type SimpleDatabaseQuerier struct {
	db *sql.DB
}

func NewSimpleDatabaseQuerier(db *sql.DB) *SimpleDatabaseQuerier {
	return &SimpleDatabaseQuerier{db: db}
}

func (s *SimpleDatabaseQuerier) Exec(query string, args ...any) (sql.Result, error) {
	return s.db.Exec(query, args...)
}

func (s *SimpleDatabaseQuerier) Query(query string, args ...any) (*sql.Rows, error) {
	return s.db.Query(query, args...)
}

func (s *SimpleDatabaseQuerier) QueryRow(query string, args ...any) *sql.Row {
	return s.db.QueryRow(query, args...)
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
