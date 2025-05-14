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

type TransactionHelper struct {
	Transaction *sql.Tx
	committed   bool
}

func NewTransaction(db *sql.DB) (*TransactionHelper, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	return &TransactionHelper{
		Transaction: tx,
	}, nil
}

func (th *TransactionHelper) Commit() error {
	if err := th.Transaction.Commit(); err != nil {
		return err
	}

	th.committed = true
	return nil
}

func (th *TransactionHelper) Rollback() error {
	if th.committed {
		return nil
	}

	if err := th.Transaction.Rollback(); err != nil {
		return err
	}

	return nil
}
