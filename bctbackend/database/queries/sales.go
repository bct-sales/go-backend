package queries

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"
	"slices"
)

// AddSale adds a sale to the database.
// A ErrSaleMissingItems is returned if itemIds is empty.
// A ErrNoSuchItem is returned if any item ID in itemIds does not correspond to any item.
// A ErrNoSuchUser is returned if the cashierId does not correspond to any user.
// A ErrSaleRequiresCashier is returned if the cashierId does not correspond to a cashier.
// A ErrDuplicateItemInSale is returned if itemIds contains duplicate item IDs.
func AddSale(
	db *sql.DB,
	cashierId models.Id,
	transactionTime models.Timestamp,
	itemIds []models.Id) (r_result models.Id, r_err error) {

	// Ensure there is at least one item in the sale.
	if len(itemIds) == 0 {
		return 0, dberr.ErrSaleMissingItems
	}

	// Check for duplicates in the item IDs.
	indexOfDuplicate := algorithms.ContainsDuplicate(itemIds)
	if indexOfDuplicate != -1 {
		duplicatedItemId := itemIds[indexOfDuplicate]
		return 0, fmt.Errorf("failed to add sale with duplicated item %d: %w", duplicatedItemId, dberr.ErrDuplicateItemInSale)
	}

	// Ensure the user exists and is a cashier
	cashier, err := GetUserWithId(db, cashierId)
	if err != nil {
		return 0, err
	}
	if !cashier.RoleId.IsCashier() {
		return 0, dberr.ErrSaleRequiresCashier
	}

	// Start a transaction
	transaction, err := NewTransaction(db)
	if err != nil {
		return 0, err
	}
	defer func() { r_err = errors.Join(r_err, transaction.Rollback()) }()

	// Check if all items exist
	exists, err := ItemsExist(transaction.transaction, itemIds)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, fmt.Errorf("failed to add sale: %w", dberr.ErrNoSuchItem)
	}

	// Check if any of the items are hidden
	if err := EnsureNoHiddenItems(transaction, itemIds); err != nil {
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
		return 0, err
	}

	saleId, err := result.LastInsertId()
	if err != nil {
		return 0, err
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
			return 0, err
		}
	}

	err = transaction.Commit()
	if err != nil {
		return 0, err
	}

	return models.Id(saleId), nil
}

type GetSalesQuery struct {
	minimalId    *models.Id // If set, only sales with an ID greater than or equal to this value are returned.
	rowSelection *struct {
		limit  int
		offset int
	}
	order *string
}

func NewGetSalesQuery() *GetSalesQuery {
	return &GetSalesQuery{}
}

func (q *GetSalesQuery) WithIdGreaterThanOrEqualTo(minimalId models.Id) *GetSalesQuery {
	q.minimalId = &minimalId
	return q
}

func (q *GetSalesQuery) WithRowSelection(limit, offset int) *GetSalesQuery {
	q.rowSelection = &struct {
		limit  int
		offset int
	}{limit: limit, offset: offset}

	return q
}

func (q *GetSalesQuery) OrderedAntiChronologically() *GetSalesQuery {
	order := "ORDER BY sales.transaction_time DESC, sales.sale_id DESC"
	q.order = &order
	return q
}

func (q *GetSalesQuery) Execute(db QueryHandler, receiver func(*models.SaleSummary) error) (r_err error) {
	query := fmt.Sprintf(
		`
			SELECT sales.sale_id, sales.cashier_id, sales.transaction_time, COUNT(sale_items.item_id) AS item_count, SUM(items.price_in_cents) AS total_price
			FROM sales
			INNER JOIN sale_items ON sales.sale_id = sale_items.sale_id
			INNER JOIN items ON sale_items.item_id = items.item_id
			%s
			GROUP BY sales.sale_id
			%s
			%s
		`,
		q.whereClause(),
		q.orderClause(),
		q.rowSelectionClause(),
	)

	queryArguments := slices.Concat(q.whereArguments())
	rows, err := db.Query(query, queryArguments...)
	if err != nil {
		return err
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	for rows.Next() {
		var saleId models.Id
		var cashierId models.Id
		var transactionTime models.Timestamp
		var itemCount int
		var totalPriceInCents models.MoneyInCents
		if err := rows.Scan(&saleId, &cashierId, &transactionTime, &itemCount, &totalPriceInCents); err != nil {
			return err
		}

		saleSummary := models.SaleSummary{
			SaleID:            saleId,
			CashierID:         cashierId,
			TransactionTime:   transactionTime,
			ItemCount:         itemCount,
			TotalPriceInCents: totalPriceInCents,
		}
		if err := receiver(&saleSummary); err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return nil
}

func (q *GetSalesQuery) whereClause() string {
	if q.minimalId == nil {
		return ""
	}
	return "WHERE sales.sale_id >= ?"
}

func (q *GetSalesQuery) whereArguments() []any {
	if q.minimalId == nil {
		return nil
	}
	return []any{*q.minimalId}
}

func (q *GetSalesQuery) rowSelectionClause() string {
	if q.rowSelection == nil {
		return ""
	}
	return fmt.Sprintf("LIMIT %d OFFSET %d", q.rowSelection.limit, q.rowSelection.offset)
}

func (q *GetSalesQuery) orderClause() string {
	if q.order == nil {
		return ""
	} else {
		return *q.order
	}
}

// GetSaleWithId returns the sale with the given saleId.
// A ErrNoSuchSale is returned if no sale with the given saleId exists.
func GetSaleWithId(db *sql.DB, saleId models.Id) (*models.Sale, error) {
	var cashierId models.Id
	var transactionTime models.Timestamp
	err := db.QueryRow(
		`
			SELECT cashier_id, transaction_time
			FROM sales
			WHERE sale_id = ?
		`,
		saleId,
	).Scan(&cashierId, &transactionTime)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to get sale with id %d: %w", saleId, dberr.ErrNoSuchSale)
	}
	if err != nil {
		return nil, err
	}

	sale := models.Sale{
		SaleID:          saleId,
		CashierID:       cashierId,
		TransactionTime: transactionTime,
	}
	return &sale, nil
}

func SaleWithIdExists(db *sql.DB, saleId models.Id) (bool, error) {
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
// Returns ErrNoSuchSale if the sale does not exist.
func GetSaleItems(db *sql.DB, saleId models.Id) (r_result []*models.Item, r_err error) {
	saleExists, err := SaleWithIdExists(db, saleId)
	if err != nil {
		return nil, err
	}
	if !saleExists {
		return nil, fmt.Errorf("failed to get items of sale %d: %w", saleId, dberr.ErrNoSuchSale)
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

	var items []*models.Item
	for rows.Next() {
		var itemId models.Id
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var categoryId models.Id
		var sellerId models.Id
		var donation bool
		var charity bool
		var frozen bool
		var hidden bool
		err := rows.Scan(&itemId, &addedAt, &description, &priceInCents, &categoryId, &sellerId, &donation, &charity, &frozen)
		if err != nil {
			return nil, err
		}

		item := models.Item{
			ItemID:       itemId,
			AddedAt:      addedAt,
			Description:  description,
			PriceInCents: priceInCents,
			CategoryID:   categoryId,
			SellerID:     sellerId,
			Donation:     donation,
			Charity:      charity,
			Frozen:       frozen,
			Hidden:       hidden,
		}
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return items, nil
}

func RemoveSale(db *sql.DB, saleId models.Id) (r_err error) {
	saleExists, err := SaleWithIdExists(db, saleId)

	if err != nil {
		return err
	}

	if !saleExists {
		return fmt.Errorf("failed to remove sale %d: %w", saleId, dberr.ErrNoSuchSale)
	}

	transaction, err := NewTransaction(db)
	if err != nil {
		return err
	}
	defer func() { r_err = errors.Join(r_err, transaction.Rollback()) }()

	_, err = transaction.Exec(
		`
			DELETE FROM sale_items
			WHERE sale_id = ?
		`,
		saleId,
	)

	if err != nil {
		return err
	}

	_, err = transaction.Exec(
		`
			DELETE FROM sales
			WHERE sale_id = ?
		`,
		saleId,
	)

	if err != nil {
		return err
	}

	err = transaction.Commit()

	if err != nil {
		return err
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
		var itemId models.Id
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var categoryId models.Id
		var sellerId models.Id
		var donation bool
		var charity bool
		var frozen bool
		var hidden bool
		err := rows.Scan(&itemId, &addedAt, &description, &priceInCents, &categoryId, &sellerId, &donation, &charity, &frozen)
		if err != nil {
			return nil, err
		}

		item := models.Item{
			ItemID:       itemId,
			AddedAt:      addedAt,
			Description:  description,
			PriceInCents: priceInCents,
			CategoryID:   categoryId,
			SellerID:     sellerId,
			Donation:     donation,
			Charity:      charity,
			Frozen:       frozen,
			Hidden:       hidden,
		}
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return items, nil
}

func GetSoldItemsCount(db QueryHandler) (r_result int, r_err error) {
	var count int
	err := db.QueryRow(
		`
			SELECT COUNT(si.item_id)
			FROM sale_items si
		`,
	).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

// HasAnyBeenSold checks if any one of the given item was involved in one or more sales.
// Does not check if items exist.
func HasAnyBeenSold(db *sql.DB, itemIds []models.Id) (r_result bool, r_err error) {
	query := fmt.Sprintf(`
		SELECT 1
		FROM items INNER JOIN sale_items ON items.item_id = sale_items.item_id
		WHERE items.item_id IN (%s)
	`, placeholderString(len(itemIds)))
	convertedItemIds := algorithms.Map(itemIds, func(id models.Id) any { return id })

	rows, err := db.Query(query, convertedItemIds...)
	if err != nil {
		return false, err
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	count := 0
	for rows.Next() {
		count++
	}

	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return count > 0, nil
}

// GetItemsSoldBy returns a list of all items sold by a specified cashier.
// The items are ordered by transaction time (most recent first) and item ID (lowest first).
func GetItemsSoldBy(db *sql.DB, cashierId models.Id) (r_result []*models.Item, r_err error) {
	if err := EnsureUserExistsAndHasRole(db, cashierId, models.NewCashierRoleId()); err != nil {
		return nil, err
	}

	rows, err := db.Query(
		`
			SELECT i.item_id, i.added_at, i.description, i.price_in_cents, i.item_category_id, i.seller_id, i.donation, i.charity, i.frozen, i.hidden
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
		var itemId models.Id
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var categoryId models.Id
		var sellerId models.Id
		var donation bool
		var charity bool
		var frozen bool
		var hidden bool

		err := rows.Scan(&itemId, &addedAt, &description, &priceInCents, &categoryId, &sellerId, &donation, &charity, &frozen, &hidden)

		if err != nil {
			return nil, err
		}

		item := models.Item{
			ItemID:       itemId,
			AddedAt:      addedAt,
			Description:  description,
			PriceInCents: priceInCents,
			CategoryID:   categoryId,
			SellerID:     sellerId,
			Donation:     donation,
			Charity:      charity,
			Frozen:       frozen,
			Hidden:       hidden,
		}
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return items, nil
}

// GetSalesWithItem returns a list of the ids of all sales that include a specified item.
// The ids are returned in ascending order.
func GetSalesWithItem(db *sql.DB, itemId models.Id) (r_result []models.Id, r_err error) {
	if itemExists, err := ItemWithIdExists(db, itemId); err != nil || !itemExists {
		if !itemExists {
			return nil, fmt.Errorf("failed to get sales with item %d: %w", itemId, dberr.ErrNoSuchItem)
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

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return saleIds, nil
}

// GetSalesWithCashier returns a list of all sales made by a specified cashier.
// The sales are ordered by transaction time (chronologically) and sale ID (lowest first).
// Returns ErrNoSuchUser if the cashierId does not correspond to any user.
// Returns ErrWrongRole if the cashierId does not correspond to a cashier.
func GetSalesWithCashier(db *sql.DB, cashierId models.Id) (r_result []*models.Sale, r_err error) {
	if err := EnsureUserExistsAndHasRole(db, cashierId, models.NewCashierRoleId()); err != nil {
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
		var saleId models.Id
		var cashierId models.Id
		var transactionTime models.Timestamp
		err := rows.Scan(&cashierId, &saleId, &transactionTime)
		if err != nil {
			return nil, err
		}

		sale := models.Sale{
			SaleID:          saleId,
			CashierID:       cashierId,
			TransactionTime: transactionTime,
		}
		sales = append(sales, &sale)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return sales, nil
}

// RemoveAllSales removes all sales from the database.
func RemoveAllSales(db *sql.DB) (r_err error) {
	transaction, err := NewTransaction(db)
	if err != nil {
		return err
	}
	defer func() { r_err = errors.Join(r_err, transaction.Rollback()) }()

	_, err = transaction.Exec(
		`
			DELETE FROM sale_items
		`,
	)
	if err != nil {
		return err
	}

	_, err = transaction.Exec(
		`
			DELETE FROM sales
		`,
	)
	if err != nil {
		return err
	}

	err = transaction.Commit()
	if err != nil {
		return err
	}

	return nil
}

func GetCashierSales(db *sql.DB, cashierId models.Id, receiver func(*models.SaleSummary) error) (r_err error) {
	if err := EnsureUserExistsAndHasRole(db, cashierId, models.NewCashierRoleId()); err != nil {
		return err
	}

	rows, err := db.Query(
		`
			SELECT sales.sale_id, sales.cashier_id, sales.transaction_time, COUNT(sale_items.item_id) AS item_count, SUM(items.price_in_cents) AS total_price
			FROM sales
			INNER JOIN sale_items ON sales.sale_id = sale_items.sale_id
			INNER JOIN items ON sale_items.item_id = items.item_id
			WHERE sales.cashier_id = ?
			GROUP BY sales.sale_id
		`,
		cashierId,
	)
	if err != nil {
		return err
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	for rows.Next() {
		var saleId models.Id
		var cashierId models.Id
		var transactionTime models.Timestamp
		var itemCount int
		var totalPriceInCents models.MoneyInCents
		if err := rows.Scan(&saleId, &cashierId, &transactionTime, &itemCount, &totalPriceInCents); err != nil {
			return err
		}

		saleSummary := models.SaleSummary{
			SaleID:            saleId,
			CashierID:         cashierId,
			TransactionTime:   transactionTime,
			ItemCount:         itemCount,
			TotalPriceInCents: totalPriceInCents,
		}
		if err := receiver(&saleSummary); err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return nil
}

func GetSalesCount(db QueryHandler) (r_result int, r_err error) {
	var count int
	err := db.QueryRow(
		`
			SELECT COUNT(*)
			FROM sales
		`,
	).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetTotalSalesValue(db QueryHandler) (r_result models.MoneyInCents, r_err error) {
	var totalValue models.MoneyInCents
	err := db.QueryRow(
		`
			SELECT SUM(items.price_in_cents) as total
			FROM sales
			INNER JOIN sale_items ON sales.sale_id = sale_items.sale_id
			INNER JOIN items ON sale_items.item_id = items.item_id
		`,
	).Scan(&totalValue)

	if err != nil {
		return 0, err
	}

	return totalValue, nil
}
