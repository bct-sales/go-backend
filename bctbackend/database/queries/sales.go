package queries

import (
	"bctbackend/database/models"
	"database/sql"
	"errors"
	"log"
)

func rollbackTransaction(transaction *sql.Tx, err error) error {
	if rollbackError := transaction.Rollback(); rollbackError != nil {
		log.Fatalf("Transaction rollback failed! Transaction error: %v, rollback error: %v", err, rollbackError)
	}

	return err
}

// AddSale adds a sale to the database.
// A SaleMissingItemsError is returned if itemIds is empty.
// A NoSuchUserError is returned if the cashierId does not correspond to any user.
// A SaleRequiresCashierError is returned if the cashierId does not correspond to a cashier.
func AddSale(
	db *sql.DB,
	cashierId models.Id,
	transactionTime models.Timestamp,
	itemIds []models.Id) (models.Id, error) {

	if len(itemIds) == 0 {
		return 0, &SaleMissingItemsError{}
	}

	cashier, err := GetUserWithId(db, cashierId)

	if err != nil {
		return 0, err
	}

	if cashier.RoleId != models.CashierRoleId {
		return 0, &SaleRequiresCashierError{}
	}

	transaction, err := db.Begin()
	if err != nil {
		return 0, err
	}

	result, err := transaction.Exec(
		`
			INSERT INTO sales(cashier_id, transaction_time)
			VALUES (?, ?)
		`,
		cashierId,
		transactionTime,
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
			SELECT sale_id, cashier_id, transaction_time
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

		err := rows.Scan(&sale.SaleId, &sale.CashierId, &sale.TransactionTime)

		if err != nil {
			return nil, err
		}

		sales = append(sales, sale)
	}

	return sales, nil
}

// GetSaleWithId returns the sale with the given saleId.
// A NoSuchSaleError is returned if no sale with the given saleId exists.
func GetSaleWithId(db *sql.DB, saleId models.Id) (models.Sale, error) {
	var sale models.Sale

	err := db.QueryRow(
		`
			SELECT sale_id, cashier_id, transaction_time
			FROM sales
			WHERE sale_id = ?
		`,
		saleId,
	).Scan(&sale.SaleId, &sale.CashierId, &sale.TransactionTime)

	if errors.Is(err, sql.ErrNoRows) {
		return sale, NoSuchSaleError{SaleId: saleId}
	}

	if err != nil {
		return sale, err
	}

	return sale, nil
}

func SaleExists(db *sql.DB, saleId models.Id) (bool, error) {
	var exists int64

	err := db.QueryRow(
		`
			SELECT 1
			FROM sales
			WHERE sale_id = ?
		`,
		saleId,
	).Scan(&exists)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func GetSaleItems(db *sql.DB, saleId models.Id) ([]models.Item, error) {
	saleExists, err := SaleExists(db, saleId)

	if err != nil {
		return nil, err
	}

	if !saleExists {
		return nil, &NoSuchSaleError{SaleId: saleId}
	}

	rows, err := db.Query(
		`
			SELECT i.item_id, i.added_at, i.description, i.price_in_cents, i.item_category_id, i.seller_id, i.donation, i.charity
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

		err := rows.Scan(&item.ItemId, &item.AddedAt, &item.Description, &item.PriceInCents, &item.CategoryId, &item.SellerId, &item.Donation, &item.Charity)

		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func RemoveSale(db *sql.DB, saleId models.Id) error {
	saleExists, err := SaleExists(db, saleId)

	if err != nil {
		return err
	}

	if !saleExists {
		return NoSuchSaleError{SaleId: saleId}
	}

	transaction, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = transaction.Exec(
		`
			DELETE FROM sale_items
			WHERE sale_id = ?
		`,
		saleId,
	)

	if err != nil {
		return rollbackTransaction(transaction, err)
	}

	_, err = transaction.Exec(
		`
			DELETE FROM sales
			WHERE sale_id = ?
		`,
		saleId,
	)

	if err != nil {
		return rollbackTransaction(transaction, err)
	}

	err = transaction.Commit()

	if err != nil {
		return rollbackTransaction(transaction, err)
	}

	return nil
}
