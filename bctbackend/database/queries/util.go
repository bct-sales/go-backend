package queries

import (
	"database/sql"
	"fmt"
	"strings"
)

func placeholderString(placeholderCount int) string {
	placeholders := make([]string, placeholderCount)
	for i := range placeholderCount {
		placeholders[i] = "?"
	}

	return strings.Join(placeholders, ", ")
}

type Transaction struct {
	transaction *sql.Tx
	committed   bool
}

func NewTransaction(db *sql.DB) (*Transaction, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start new transaction: %w", err)
	}

	return &Transaction{
		transaction: tx,
	}, nil
}

func (t *Transaction) Commit() error {
	if err := t.transaction.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	t.committed = true
	return nil
}

func (t *Transaction) Rollback() error {
	if t.committed {
		return nil
	}

	if err := t.transaction.Rollback(); err != nil {
		return fmt.Errorf("failed to roll back transaction: %w", err)
	}

	return nil
}

func (t *Transaction) Exec(query string, args ...any) (sql.Result, error) {
	return t.transaction.Exec(query, args...)
}

func (t *Transaction) Query(query string, args ...any) (*sql.Rows, error) {
	return t.transaction.Query(query, args...)
}

func (t *Transaction) QueryRow(query string, args ...any) *sql.Row {
	return t.transaction.QueryRow(query, args...)
}

type SQLOption interface {
	SQL() string
}

type allRows struct{}

func (*allRows) SQL() string {
	return ""
}

type rowSelection struct {
	Limit  int
	Offset int
}

func (p *rowSelection) SQL() string {
	return fmt.Sprintf("LIMIT %d OFFSET %d", p.Limit, p.Offset)
}

func AllRows() SQLOption {
	return &allRows{}
}

func RowSelection(offset int, limit int) SQLOption {
	if limit < 0 || offset < 0 {
		panic("limit and offset must be non-negative")
	}

	return &rowSelection{Limit: limit, Offset: offset}
}
