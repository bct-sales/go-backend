package queries

import (
	"database/sql"
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
		return nil, err
	}

	return &Transaction{
		transaction: tx,
	}, nil
}

func (t *Transaction) Commit() error {
	if err := t.transaction.Commit(); err != nil {
		return err
	}

	t.committed = true
	return nil
}

func (t *Transaction) Rollback() error {
	if t.committed {
		return nil
	}

	if err := t.transaction.Rollback(); err != nil {
		return err
	}

	return nil
}

func (t *Transaction) Exec(query string, args ...any) (sql.Result, error) {
	return t.transaction.Exec(query, args...)
}

func (t *Transaction) Query(query string, args ...any) (*sql.Rows, error) {
	return t.transaction.Query(query, args...)
}
