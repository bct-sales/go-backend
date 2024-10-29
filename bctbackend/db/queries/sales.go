package queries

import (
	models "bctbackend/db/models"
	"database/sql"
	"log"
)

func rollbackTransaction(transaction *sql.Tx, err error) error {
	if rollbackError := transaction.Rollback(); rollbackError != nil {
		log.Fatalf("Transaction rollback failed! Transaction error: %v, rollback error: %v", err, rollbackError)
	}

	return err
}

func AddSale(
	db *sql.DB,
	cashierId models.Id,
	timestamp models.Timestamp,
	itemIds []models.Id) (models.Id, error) {

	transaction, err := db.Begin()
	if err != nil {
		return 0, err
	}

	result, err := transaction.Exec(
		`
			INSERT INTO sales(cashier_id, timestamp)
			VALUES (?, ?)
		`,
		cashierId,
		timestamp,
	)

	if err != nil {
		return 0, rollbackTransaction(transaction, err)
	}

	saleId, err := result.LastInsertId()

	if err != nil {
		return 0, rollbackTransaction(transaction, err)
	}

	for _, itemId := range itemIds {
		_, err := transaction.Exec(
			`
				INSERT INTO sale_items(sale_id, item_id)
				VALUES (?, ?)
			`,
			saleId,
			itemId,
		)

		if err != nil {
			return 0, rollbackTransaction(transaction, err)
		}
	}

	err = transaction.Commit()

	if err != nil {
		return 0, rollbackTransaction(transaction, err)
	}

	return saleId, nil
}
