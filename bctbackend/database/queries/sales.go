package queries

import (
	"bctbackend/algorithms"
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
// An NoSuchItemError is returned if any item ID in itemIds does not correspond to any item.
// A NoSuchUserError is returned if the cashierId does not correspond to any user.
// A SaleRequiresCashierError is returned if the cashierId does not correspond to a cashier.
// A DuplicateItemInSaleError is returned if itemIds contains duplicate item IDs.
func AddSale(
	db *sql.DB,
	cashierId models.Id,
	transactionTime models.Timestamp,
	itemIds []models.Id) (models.Id, error) {

	// Ensure there is at least one item in the sale.
	if len(itemIds) == 0 {
		return 0, &SaleMissingItemsError{}
	}

	// Check for duplicates in the item IDs.
	indexOfDuplicate := algorithms.ContainsDuplicate(itemIds)
	if indexOfDuplicate != -1 {
		duplicatedItemId := itemIds[indexOfDuplicate]
		return 0, &DuplicateItemInSaleError{ItemId: duplicatedItemId}
	}

	// Ensure the user exists and is a cashier.
	cashier, err := GetUserWithId(db, cashierId)
	if err != nil {
		return 0, err
	}
	if cashier.RoleId != models.CashierRoleId {
		return 0, &SaleRequiresCashierError{}
	}

	// Ensure that all items exists
	for _, itemId := range itemIds {
		itemExists, err := ItemWithIdExists(db, itemId)
		if err != nil {
			return 0, err
		}
		if !itemExists {
			return 0, &NoSuchItemError{Id: itemId}
		}
	}

	// Start a transaction.
	transaction, err := db.Begin()
	if err != nil {
		return 0, err
	}

	// Create sale
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

	// Add items to sale
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

func GetSales(db *sql.DB, receiver func(*models.SaleSummary) error) (r_err error) {
	rows, err := db.Query(
		`
			SELECT sales.sale_id, sales.cashier_id, sales.transaction_time, COUNT(sale_items.item_id) AS item_count, SUM(items.price_in_cents) AS total_price
			FROM sales
			INNER JOIN sale_items ON sales.sale_id = sale_items.sale_id
			INNER JOIN items ON sale_items.item_id = items.item_id
			GROUP BY sales.sale_id
		`,
	)
	if err != nil {
		return err
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	for rows.Next() {
		var sale models.SaleSummary

		if err := rows.Scan(&sale.SaleId, &sale.CashierId, &sale.TransactionTime, &sale.ItemCount, &sale.TotalPriceInCents); err != nil {
			return err
		}

		if err := receiver(&sale); err != nil {
			return err
		}
	}

	return nil
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
		return sale, &NoSuchSaleError{SaleId: saleId}
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

// GetSaleItems lists all items associated with a specified sale.
func GetSaleItems(db *sql.DB, saleId models.Id) (r_result []models.Item, r_err error) {
	saleExists, err := SaleExists(db, saleId)

	if err != nil {
		return nil, err
	}

	if !saleExists {
		return nil, &NoSuchSaleError{SaleId: saleId}
	}

	rows, err := db.Query(
		`
			SELECT i.item_id, i.added_at, i.description, i.price_in_cents, i.item_category_id, i.seller_id, i.donation, i.charity, i.frozen
			FROM sale_items si
			INNER JOIN items i ON si.item_id = i.item_id
			WHERE si.sale_id = ?
		`,
		saleId,
	)

	if err != nil {
		return nil, err
	}

	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	var items []models.Item
	for rows.Next() {
		var item models.Item

		err := rows.Scan(&item.ItemId, &item.AddedAt, &item.Description, &item.PriceInCents, &item.CategoryId, &item.SellerId, &item.Donation, &item.Charity, &item.Frozen)

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
		return &NoSuchSaleError{SaleId: saleId}
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

// GetSoldItems returns a list of all items that have been sold.
// The items are ordered by transaction time (most recent first) and item ID (lowest first).
func GetSoldItems(db *sql.DB) (r_result []*models.Item, r_err error) {
	rows, err := db.Query(
		`
			SELECT DISTINCT i.item_id, i.added_at, i.description, i.price_in_cents, i.item_category_id, i.seller_id, i.donation, i.charity, i.frozen
			FROM sale_items si
			INNER JOIN items i ON si.item_id = i.item_id
			INNER JOIN sales s ON si.sale_id = s.sale_id
			ORDER BY s.transaction_time DESC, i.item_id ASC
		`,
	)

	if err != nil {
		return nil, err
	}

	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	var items []*models.Item
	for rows.Next() {
		var item models.Item

		err := rows.Scan(&item.ItemId, &item.AddedAt, &item.Description, &item.PriceInCents, &item.CategoryId, &item.SellerId, &item.Donation, &item.Charity, &item.Frozen)

		if err != nil {
			return nil, err
		}

		items = append(items, &item)
	}

	return items, nil
}

// GetItemsSoldBy returns a list of all items sold by a specified cashier.
// The items are ordered by transaction time (most recent first) and item ID (lowest first).
func GetItemsSoldBy(db *sql.DB, cashierId models.Id) (r_result []*models.Item, r_err error) {
	if err := CheckUserRole(db, cashierId, models.CashierRoleId); err != nil {
		return nil, err
	}

	rows, err := db.Query(
		`
			SELECT i.item_id, i.added_at, i.description, i.price_in_cents, i.item_category_id, i.seller_id, i.donation, i.charity, i.frozen
			FROM sale_items si
			INNER JOIN items i ON si.item_id = i.item_id
			INNER JOIN sales s ON si.sale_id = s.sale_id
			WHERE s.cashier_id = ?
			ORDER BY s.transaction_time DESC, i.item_id ASC
		`,
		cashierId,
	)

	if err != nil {
		return nil, err
	}

	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	var items []*models.Item
	for rows.Next() {
		var item models.Item

		err := rows.Scan(&item.ItemId, &item.AddedAt, &item.Description, &item.PriceInCents, &item.CategoryId, &item.SellerId, &item.Donation, &item.Charity, &item.Frozen)

		if err != nil {
			return nil, err
		}

		items = append(items, &item)
	}

	return items, nil
}

// GetSalesWithItem returns a list of the ids of all sales that include a specified item.
// The ids are returned in ascending order.
func GetSalesWithItem(db *sql.DB, itemId models.Id) (r_result []models.Id, r_err error) {
	if itemExists, err := ItemWithIdExists(db, itemId); err != nil || !itemExists {
		if !itemExists {
			return nil, &NoSuchItemError{Id: itemId}
		}

		return nil, err
	}

	rows, err := db.Query(
		`
			SELECT sale_id
			FROM sale_items
			WHERE item_id = ?
			ORDER BY sale_id ASC
		`,
		itemId,
	)

	if err != nil {
		return nil, err
	}

	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	saleIds := []models.Id{}
	for rows.Next() {
		var saleId models.Id

		err := rows.Scan(&saleId)

		if err != nil {
			return nil, err
		}

		saleIds = append(saleIds, saleId)
	}

	return saleIds, nil
}

// GetSalesWithCashier returns a list of all sales made by a specified cashier.
// The sales are ordered by transaction time (chronologically) and sale ID (lowest first).
// Returns NoSuchUserError if the cashierId does not correspond to any user.
// Returns InvalidRoleError if the cashierId does not correspond to a cashier.
func GetSalesWithCashier(db *sql.DB, cashierId models.Id) (r_result []*models.Sale, r_err error) {
	if err := CheckUserRole(db, cashierId, models.CashierRoleId); err != nil {
		return nil, err
	}

	rows, err := db.Query(
		`
			SELECT cashier_id, sale_id, transaction_time
			FROM sales
			WHERE cashier_id = ?
			ORDER BY transaction_time ASC, sale_id ASC
		`,
		cashierId,
	)

	if err != nil {
		return nil, err
	}

	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	sales := []*models.Sale{}
	for rows.Next() {
		var sale models.Sale

		err := rows.Scan(&sale.CashierId, &sale.SaleId, &sale.TransactionTime)

		if err != nil {
			return nil, err
		}

		sales = append(sales, &sale)
	}

	return sales, nil
}
