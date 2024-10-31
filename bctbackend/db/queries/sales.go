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

func GetSales(db *sql.DB) ([]models.Sale, error) {
	rows, err := db.Query(
		`
			SELECT sale_id, cashier_id, timestamp
			FROM sales
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var sales []models.Sale

	for rows.Next() {
		var sale models.Sale

		err := rows.Scan(&sale.SaleId, &sale.CashierId, &sale.Timestamp)

		if err != nil {
			return nil, err
		}

		sales = append(sales, sale)
	}

	return sales, nil
}

func GetSaleItems(db *sql.DB, saleId models.Id) ([]models.Item, error) {
	rows, err := db.Query(
		`
			SELECT i.item_id, i.timestamp, i.description, i.price_in_cents, i.item_category_id, i.seller_id, i.donation, i.charity
			FROM sale_items si
			INNER JOIN items i ON si.item_id = i.item_id
			WHERE si.sale_id = ?
		`,
		saleId,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var items []models.Item

	for rows.Next() {
		var item models.Item

		err := rows.Scan(&item.ItemId, &item.Timestamp, &item.Description, &item.PriceInCents, &item.CategoryId, &item.SellerId, &item.Donation, &item.Charity)

		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}