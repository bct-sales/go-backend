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

type addSaleQuery struct {
	CashierID       models.ID
	TransactionTime models.Timestamp
	ItemIDs         []models.ID
}

func (q *addSaleQuery) execute(db *TransactionalDatabaseQuerier) (r_result models.ID, r_err error) {
	if err := q.ensureInputsValidity(db); err != nil {
		return 0, err
	}

	// Check if all items exist
	if err := EnsureItemsExist(db, q.ItemIDs); err != nil {
		return 0, fmt.Errorf("failed to add sale: %w", dberr.ErrNoSuchItem)
	}

	// Check if any of the items are hidden
	if err := EnsureNoHiddenItems(db, q.ItemIDs); err != nil {
		return 0, err
	}

	// Create sale
	result, err := db.Exec(
		`
			INSERT INTO sales(cashier_id, transaction_time)
			VALUES (?, ?)
		`,
		q.CashierID,
		q.TransactionTime,
	)
	if err != nil {
		return 0, err
	}

	saleID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Add items to sale
	for _, itemID := range q.ItemIDs {
		_, err := db.Exec(
			`
				INSERT INTO sale_items(sale_id, item_id)
				VALUES (?, ?)
			`,
			saleID,
			itemID,
		)

		if err != nil {
			return 0, err
		}
	}

	return models.ID(saleID), nil
}

func (q *addSaleQuery) ensureInputsValidity(db DatabaseQuerier) error {
	// Ensure there is at least one item in the sale.
	if len(q.ItemIDs) == 0 {
		return dberr.ErrSaleMissingItems
	}

	// Check for duplicates in the item IDs.
	indexOfDuplicate := algorithms.ContainsDuplicate(q.ItemIDs)
	if indexOfDuplicate != -1 {
		duplicatedItemID := q.ItemIDs[indexOfDuplicate]
		return fmt.Errorf("failed to add sale with duplicated item %d: %w", duplicatedItemID, dberr.ErrDuplicateItemInSale)
	}

	// Ensure the user exists and is a cashier
	cashier, err := GetUserWithID(db, q.CashierID)
	if err != nil {
		return err
	}
	if !cashier.RoleID.IsCashier() {
		return dberr.ErrSaleRequiresCashier
	}

	return nil
}

// AddSale adds a sale to the database.
// A ErrSaleMissingItems is returned if itemIDs is empty.
// A ErrNoSuchItem is returned if any item ID in itemIDs does not correspond to any item.
// A ErrNoSuchUser is returned if the cashierID does not correspond to any user.
// A ErrSaleRequiresCashier is returned if the cashierID does not correspond to a cashier.
// A ErrDuplicateItemInSale is returned if itemIDs contains duplicate item IDs.
func AddSale(
	db *TransactionalDatabaseQuerier,
	cashierID models.ID,
	transactionTime models.Timestamp,
	itemIDs []models.ID) (r_result models.ID, r_err error) {

	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	query := addSaleQuery{
		CashierID:       cashierID,
		TransactionTime: transactionTime,
		ItemIDs:         itemIDs,
	}

	return query.execute(db)
}

type GetSalesQuery struct {
	minimalID    *models.ID // If set, only sales with an ID greater than or equal to this value are returned.
	order        *string    // Specifies the order in which to return the results
	rowSelection RowRange
}

func NewGetSalesQuery() *GetSalesQuery {
	return &GetSalesQuery{
		minimalID:    nil,
		rowSelection: RowRange{},
		order:        nil,
	}
}

func (q *GetSalesQuery) WithIDGreaterThanOrEqualTo(minimalID models.ID) *GetSalesQuery {
	q.minimalID = &minimalID
	return q
}

func (q *GetSalesQuery) WithRowRange(limit uint64, offset uint64) *GetSalesQuery {
	q.rowSelection = RowRange{Limit: &limit, Offset: &offset}

	return q
}

func (q *GetSalesQuery) OrderedAntiChronologically() *GetSalesQuery {
	order := "ORDER BY sales.transaction_time DESC, sales.sale_id DESC"
	q.order = &order
	return q
}

func (q *GetSalesQuery) Execute(db DatabaseQuerier, receiver func(*models.SaleSummary) error) (r_err error) {
	query := fmt.Sprintf(
		`
			SELECT
				sales.sale_id,
				sales.cashier_id,
				sales.transaction_time,
				COUNT(sale_items.item_id) AS item_count,
				SUM(items.price_in_cents) AS total_price
			FROM
				sales
			INNER JOIN
				sale_items ON sales.sale_id = sale_items.sale_id
			INNER JOIN
				items ON sale_items.item_id = items.item_id
			%s
			GROUP BY
				sales.sale_id
			%s
			%s
		`,
		q.whereClause(),
		q.orderClause(),
		q.rowSelection.SQL(),
	)

	queryArguments := slices.Concat(q.whereArguments())
	rows, err := db.Query(query, queryArguments...)
	if err != nil {
		return err
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	for rows.Next() {
		var saleID models.ID
		var cashierID models.ID
		var transactionTime models.Timestamp
		var itemCount int
		var totalPriceInCents models.MoneyInCents
		if err := rows.Scan(&saleID, &cashierID, &transactionTime, &itemCount, &totalPriceInCents); err != nil {
			return err
		}

		saleSummary := models.SaleSummary{
			SaleID:            saleID,
			CashierID:         cashierID,
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
	if q.minimalID == nil {
		return ""
	}
	return "WHERE sales.sale_id >= ?"
}

func (q *GetSalesQuery) whereArguments() []any {
	if q.minimalID == nil {
		return nil
	}
	return []any{*q.minimalID}
}

func (q *GetSalesQuery) orderClause() string {
	if q.order == nil {
		return ""
	} else {
		return *q.order
	}
}

// GetSaleWithID returns the sale with the given saleID.
// A ErrNoSuchSale is returned if no sale with the given saleID exists.
func GetSaleWithID(db DatabaseQuerier, saleID models.ID) (r_result *models.Sale, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	var cashierID models.ID
	var transactionTime models.Timestamp
	err := db.QueryRow(
		`
			SELECT
				cashier_id,
				transaction_time
			FROM
				sales
			WHERE
				sale_id = ?
		`,
		saleID,
	).Scan(&cashierID, &transactionTime)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to get sale with id %d: %w", saleID, dberr.ErrNoSuchSale)
	}
	if err != nil {
		return nil, err
	}

	sale := models.Sale{
		SaleID:          saleID,
		CashierID:       cashierID,
		TransactionTime: transactionTime,
	}
	return &sale, nil
}

// SaleWithIDExists checks whether a sale with the given saleID exists in the database.
// Returns true if the sale exists, false otherwise.
func SaleWithIDExists(db DatabaseQuerier, saleID models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	var exists int64

	err := db.QueryRow(
		`
			SELECT
				1
			FROM
				sales
			WHERE
				sale_id = ?
		`,
		saleID,
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
func GetSaleItems(db DatabaseQuerier, saleID models.ID) (r_result []*models.Item, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	saleExists, err := SaleWithIDExists(db, saleID)
	if err != nil {
		return nil, err
	}
	if !saleExists {
		return nil, fmt.Errorf("failed to get items of sale %d: %w", saleID, dberr.ErrNoSuchSale)
	}

	rows, err := db.Query(
		`
			SELECT
				i.item_id,
				i.added_at,
				i.description,
				i.price_in_cents,
				i.item_category_id,
				i.seller_id,
				i.donation,
				i.charity,
				i.frozen,
				i.large
			FROM
				sale_items si
			INNER JOIN
				items i ON si.item_id = i.item_id
			WHERE
				si.sale_id = ?
		`,
		saleID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	var items []*models.Item
	for rows.Next() {
		var itemID models.ID
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var categoryID models.ID
		var sellerID models.ID
		var donation bool
		var charity bool
		var frozen bool
		var hidden bool
		var large bool
		err := rows.Scan(&itemID, &addedAt, &description, &priceInCents, &categoryID, &sellerID, &donation, &charity, &frozen, &large)
		if err != nil {
			return nil, err
		}

		item := models.Item{
			ItemID:       itemID,
			AddedAt:      addedAt,
			Description:  description,
			PriceInCents: priceInCents,
			CategoryID:   categoryID,
			SellerID:     sellerID,
			Donation:     donation,
			Charity:      charity,
			Frozen:       frozen,
			Hidden:       hidden,
			Large:        large,
		}
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return items, nil
}

// RemoveSale removes the sale with the given saleID from the database.
// Returns ErrNoSuchSale if no such sale exists.
// This function should be use with care.
func RemoveSale(transaction *TransactionalDatabaseQuerier, saleID models.ID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	saleExists, err := SaleWithIDExists(transaction, saleID)
	if err != nil {
		return err
	}
	if !saleExists {
		return fmt.Errorf("failed to remove sale %d: %w", saleID, dberr.ErrNoSuchSale)
	}

	_, err = transaction.Exec(
		`
			DELETE FROM sale_items
			WHERE sale_id = ?
		`,
		saleID,
	)

	if err != nil {
		return err
	}

	_, err = transaction.Exec(
		`
			DELETE FROM sales
			WHERE sale_id = ?
		`,
		saleID,
	)
	if err != nil {
		return err
	}

	return nil
}

// GetSoldItems returns a list of all items that have been sold.
// The items are ordered by transaction time (most recent first) and item ID (lowest first).
func GetSoldItems(db DatabaseQuerier) (r_result []*models.Item, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	rows, err := db.Query(
		`
			SELECT DISTINCT
				i.item_id,
				i.added_at,
				i.description,
				i.price_in_cents,
				i.item_category_id,
				i.seller_id,
				i.donation,
				i.charity,
				i.frozen
			FROM
				sale_items si
			INNER JOIN
				items i ON si.item_id = i.item_id
			INNER JOIN
				sales s ON si.sale_id = s.sale_id
			ORDER BY
				s.transaction_time DESC, i.item_id ASC
		`,
	)
	if err != nil {
		return nil, err
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	var items []*models.Item
	for rows.Next() {
		var itemID models.ID
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var categoryID models.ID
		var sellerID models.ID
		var donation bool
		var charity bool
		var frozen bool
		var hidden bool

		err := rows.Scan(
			&itemID,
			&addedAt,
			&description,
			&priceInCents,
			&categoryID,
			&sellerID,
			&donation,
			&charity,
			&frozen,
		)
		if err != nil {
			return nil, err
		}

		item := models.Item{
			ItemID:       itemID,
			AddedAt:      addedAt,
			Description:  description,
			PriceInCents: priceInCents,
			CategoryID:   categoryID,
			SellerID:     sellerID,
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

// CountSoldItems returns the total number of items that have been sold.
// Two counts are returned: one where each item is counted only once, and one where each item is counted for each sale it was involved in.
func CountSoldItems(db DatabaseQuerier) (r_result *struct {
	Distinct         int
	IncludeMultiples int
}, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	var totalCount int
	var distinctCount int
	err := db.QueryRow(
		`
			SELECT
				COUNT(si.item_id),
				COUNT(DISTINCT si.item_id)
			FROM
				sale_items si
		`,
	).Scan(&totalCount, &distinctCount)

	if err != nil {
		return nil, err
	}

	result := struct {
		Distinct         int
		IncludeMultiples int
	}{
		Distinct:         distinctCount,
		IncludeMultiples: totalCount,
	}
	return &result, nil
}

// HasAnyBeenSold checks if any one of the given item was involved in one or more sales.
// Does not check if items exist.
func HasAnyBeenSold(db DatabaseQuerier, itemIDs []models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	query := fmt.Sprintf(`
		SELECT
			1
		FROM
			items
		INNER JOIN
			sale_items ON items.item_id = sale_items.item_id
		WHERE
			items.item_id IN (%s)
	`, placeholderString(len(itemIDs)))
	convertedItemIDs := algorithms.Map(itemIDs, func(id models.ID) any { return id })

	rows, err := db.Query(query, convertedItemIDs...)
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
func GetItemsSoldBy(db DatabaseQuerier, cashierID models.ID) (r_result []*models.Item, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if err := EnsureUserExistsAndHasRole(db, cashierID, models.NewCashierRoleID()); err != nil {
		return nil, err
	}

	rows, err := db.Query(
		`
			SELECT
				i.item_id,
				i.added_at,
				i.description,
				i.price_in_cents,
				i.item_category_id,
				i.seller_id,
				i.donation,
				i.charity,
				i.frozen,
				i.hidden,
				i.large
			FROM
				sale_items si
			INNER JOIN
				items i ON si.item_id = i.item_id
			INNER JOIN
				sales s ON si.sale_id = s.sale_id
			WHERE
				s.cashier_id = ?
			ORDER BY
				s.transaction_time DESC, i.item_id ASC
		`,
		cashierID,
	)

	if err != nil {
		return nil, err
	}

	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	var items []*models.Item
	for rows.Next() {
		var itemID models.ID
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var categoryID models.ID
		var sellerID models.ID
		var donation bool
		var charity bool
		var frozen bool
		var hidden bool
		var large bool

		err := rows.Scan(&itemID, &addedAt, &description, &priceInCents, &categoryID, &sellerID, &donation, &charity, &frozen, &hidden, &large)

		if err != nil {
			return nil, err
		}

		item := models.Item{
			ItemID:       itemID,
			AddedAt:      addedAt,
			Description:  description,
			PriceInCents: priceInCents,
			CategoryID:   categoryID,
			SellerID:     sellerID,
			Donation:     donation,
			Charity:      charity,
			Frozen:       frozen,
			Hidden:       hidden,
			Large:        large,
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
func GetSalesWithItem(db DatabaseQuerier, itemID models.ID) (r_result []models.ID, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if itemExists, err := ItemWithIDExists(db, itemID); err != nil || !itemExists {
		if !itemExists {
			return nil, fmt.Errorf("failed to get sales with item %d: %w", itemID, dberr.ErrNoSuchItem)
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
		itemID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	saleIDs := []models.ID{}
	for rows.Next() {
		var saleID models.ID

		err := rows.Scan(&saleID)

		if err != nil {
			return nil, err
		}

		saleIDs = append(saleIDs, saleID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return saleIDs, nil
}

// GetSalesWithCashier returns a list of all sales made by a specified cashier.
// The sales are ordered by transaction time (chronologically) and sale ID (lowest first).
// Returns ErrNoSuchUser if the cashierID does not correspond to any user.
// Returns ErrWrongRole if the cashierID does not correspond to a cashier.
func GetSalesWithCashier(db DatabaseQuerier, cashierID models.ID) (r_result []*models.Sale, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if err := EnsureUserExistsAndHasRole(db, cashierID, models.NewCashierRoleID()); err != nil {
		return nil, err
	}

	rows, err := db.Query(
		`
			SELECT cashier_id, sale_id, transaction_time
			FROM sales
			WHERE cashier_id = ?
			ORDER BY transaction_time ASC, sale_id ASC
		`,
		cashierID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	sales := []*models.Sale{}
	for rows.Next() {
		var saleID models.ID
		var cashierID models.ID
		var transactionTime models.Timestamp
		err := rows.Scan(&cashierID, &saleID, &transactionTime)
		if err != nil {
			return nil, err
		}

		sale := models.Sale{
			SaleID:          saleID,
			CashierID:       cashierID,
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
func RemoveAllSales(transaction *TransactionalDatabaseQuerier) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	{
		_, err := transaction.Exec(
			`
			DELETE FROM sale_items
		`,
		)
		if err != nil {
			return err
		}
	}

	{
		_, err := transaction.Exec(
			`
			DELETE FROM sales
		`,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetCashierSales retrieves a list of sales made by a specified cashier.
// If cashierID does not correspond to any user, ErrNoSuchUser is returned.
// If cashierID does not correspond to a cashier, ErrWrongRole is returned.
func GetCashierSales(db DatabaseQuerier, cashierID models.ID, receiver func(*models.SaleSummary) error, order Order, rowSelection *RowRange) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if err := EnsureUserExistsAndHasRole(db, cashierID, models.NewCashierRoleID()); err != nil {
		return err
	}

	orderClause := ""
	switch order {
	case OrderChronological:
		orderClause = "ORDER BY sales.transaction_time ASC, sales.sale_id ASC"
	case OrderAntiChronological:
		orderClause = "ORDER BY sales.transaction_time DESC, sales.sale_id DESC"
	}

	query := fmt.Sprintf(
		`
			SELECT
				sales.sale_id,
				sales.cashier_id,
				sales.transaction_time,
				COUNT(sale_items.item_id) AS item_count,
				SUM(items.price_in_cents) AS total_price
			FROM
				sales
			INNER JOIN
				sale_items ON sales.sale_id = sale_items.sale_id
			INNER JOIN
				items ON sale_items.item_id = items.item_id
			WHERE
				sales.cashier_id = ?
			GROUP BY
				sales.sale_id
			%s
			%s
		`,
		orderClause,
		rowSelection.SQL(),
	)
	rows, err := db.Query(query, cashierID)
	if err != nil {
		return err
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	for rows.Next() {
		var saleID models.ID
		var cashierID models.ID
		var transactionTime models.Timestamp
		var itemCount int
		var totalPriceInCents models.MoneyInCents
		if err := rows.Scan(&saleID, &cashierID, &transactionTime, &itemCount, &totalPriceInCents); err != nil {
			return err
		}

		saleSummary := models.SaleSummary{
			SaleID:            saleID,
			CashierID:         cashierID,
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

// CountSales returns the total number of sales in the database.
func CountSales(db DatabaseQuerier) (r_result int, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

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

// CountCashierSales returns the total number of sales made by the specified cashier.
func CountCashierSales(db DatabaseQuerier, cashierID models.ID) (r_result int, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if err := EnsureUserExistsAndHasRole(db, cashierID, models.NewCashierRoleID()); err != nil {
		return 0, err
	}

	var count int
	err := db.QueryRow(
		`
			SELECT COUNT(*)
			FROM sales
			WHERE cashier_id = ?
		`,
		cashierID,
	).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetTotalSalesValue returns the total value of all sales in the database.
// The value is calculated as the sum of the prices of all items sold.
// If an item was sold multiple times, its price is counted each time.
func GetTotalSalesValue(db DatabaseQuerier) (r_result models.MoneyInCents, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	var totalValue models.MoneyInCents
	err := db.QueryRow(
		`
			SELECT
				COALESCE(SUM(items.price_in_cents), 0) as total
			FROM
				sales
			INNER JOIN
				sale_items ON sales.sale_id = sale_items.sale_id
			INNER JOIN
				items ON sale_items.item_id = items.item_id
		`,
	).Scan(&totalValue)

	if err != nil {
		return 0, err
	}

	return totalValue, nil
}

type MultiplySoldItem struct {
	Item  models.Item
	Sales []models.Sale
}

// GetMultiplySoldItems retrieves all items that have been sold multiple times.
func GetMultiplySoldItems(db DatabaseQuerier) (r_result []MultiplySoldItem, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	rows, err := db.Query(
		`
			SELECT
				item.item_id,
				item.added_at,
			    item.description,
				item.price_in_cents,
				item.item_category_id,
				item.seller_id,
				item.donation,
				item.charity,
				item.frozen,
				sale.sale_id,
				sale.cashier_id,
				sale.transaction_time
			FROM
				items item
			INNER JOIN
				item_categories category ON item.item_category_id = category.item_category_id
			INNER JOIN
				sale_items sale_item ON item.item_id = sale_item.item_id
			INNER JOIN
				sales sale ON sale_item.sale_id = sale.sale_id
			WHERE
				(
					SELECT COUNT(*)
					FROM sale_items si
					WHERE si.item_id = item.item_id
				) > 1
			ORDER BY
				item.item_id, sale.sale_id
		`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	var multiplySoldItems []MultiplySoldItem

	for rows.Next() {
		var rowData struct {
			ItemID          models.ID
			AddedAt         models.Timestamp
			Description     string
			PriceInCents    models.MoneyInCents
			CategoryID      models.ID
			SellerID        models.ID
			Donation        bool
			Charity         bool
			Frozen          bool
			SaleID          models.ID
			CashierID       models.ID
			TransactionTime models.Timestamp
		}
		err := rows.Scan(
			&rowData.ItemID,
			&rowData.AddedAt,
			&rowData.Description,
			&rowData.PriceInCents,
			&rowData.CategoryID,
			&rowData.SellerID,
			&rowData.Donation,
			&rowData.Charity,
			&rowData.Frozen,
			&rowData.SaleID,
			&rowData.CashierID,
			&rowData.TransactionTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		sale := models.Sale{
			SaleID:          rowData.SaleID,
			CashierID:       rowData.CashierID,
			TransactionTime: rowData.TransactionTime,
		}

		lastMultiplySoldItemIndex := len(multiplySoldItems) - 1

		if lastMultiplySoldItemIndex >= 0 && multiplySoldItems[lastMultiplySoldItemIndex].Item.ItemID == rowData.ItemID {
			multiplySoldItems[lastMultiplySoldItemIndex].Sales = append(multiplySoldItems[lastMultiplySoldItemIndex].Sales, sale)
		} else {
			multiplySoldItem := MultiplySoldItem{
				Item: models.Item{
					ItemID:       rowData.ItemID,
					AddedAt:      rowData.AddedAt,
					Description:  rowData.Description,
					PriceInCents: rowData.PriceInCents,
					CategoryID:   rowData.CategoryID,
					SellerID:     rowData.SellerID,
					Donation:     rowData.Donation,
					Charity:      rowData.Charity,
					Frozen:       rowData.Frozen,
					Hidden:       false,
				},
				Sales: []models.Sale{sale},
			}

			multiplySoldItems = append(multiplySoldItems, multiplySoldItem)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return multiplySoldItems, nil
}

type SaleItemInformation struct {
	SellerID       models.ID
	Description    string
	ItemCategoryID models.ID
	PriceInCents   models.MoneyInCents
	SellCount      int64
}

// GetSaleItemInformation retrieves information about a sale item.
// If the item is not found, it returns an ErrNoSuchItem.
func GetSaleItemInformation(
	db DatabaseQuerier,
	itemID models.ID) (r_result *SaleItemInformation, r_err error) {

	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	row := db.QueryRow(
		`
			SELECT seller_id, description, price_in_cents, item_category_id, COUNT(si.sale_id)
			FROM items i LEFT JOIN sale_items si ON i.item_id = si.item_id
			GROUP BY i.item_id
			HAVING i.item_id = ?
		`,
		itemID)

	var sellerID models.ID
	var description string
	var itemCategoryID models.ID
	var priceInCents models.MoneyInCents
	var sellCount int64
	err := row.Scan(
		&sellerID,
		&description,
		&priceInCents,
		&itemCategoryID,
		&sellCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to get information about item %d: %w", itemID, dberr.ErrNoSuchItem)
	}
	if err != nil {
		return nil, err
	}

	saleItemInformation := SaleItemInformation{
		SellerID:       sellerID,
		Description:    description,
		ItemCategoryID: itemCategoryID,
		PriceInCents:   priceInCents,
		SellCount:      sellCount,
	}

	return &saleItemInformation, nil
}

type CategorySaleTotal struct {
	CategoryID   models.ID
	CategoryName string
	TotalInCents models.MoneyInCents
}

// GetSalesOverview retrieves the total sales for each item category.
// Multiply sold items are counted as many times.
func GetSalesOverview(db DatabaseQuerier) (r_result []CategorySaleTotal, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	rows, err := db.Query(
		`
			SELECT item_categories.item_category_id, item_categories.name, SUM(COALESCE(i.price_in_cents, 0))
			FROM item_categories
			LEFT JOIN (
				items INNER JOIN sale_items ON items.item_id = sale_items.item_id
			) AS i ON i.item_category_id = item_categories.item_category_id
			GROUP BY item_categories.item_category_id
			ORDER BY item_categories.item_category_id
		`,
	)

	if err != nil {
		return nil, err
	}

	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	var categorySaleTotals []CategorySaleTotal

	for rows.Next() {
		var categoryID models.ID
		var categoryName string
		var totalInCents models.MoneyInCents
		err := rows.Scan(
			&categoryID,
			&categoryName,
			&totalInCents,
		)
		if err != nil {
			return nil, err
		}

		categorySaleTotal := CategorySaleTotal{
			CategoryID:   categoryID,
			CategoryName: categoryName,
			TotalInCents: totalInCents,
		}

		categorySaleTotals = append(categorySaleTotals, categorySaleTotal)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return categorySaleTotals, nil
}

type SoldItem struct {
	SaleID          models.ID
	CashierID       models.ID
	TransactionTime models.Timestamp
	ItemID          models.ID
	AddedAt         models.Timestamp
	Description     string
	PriceInCents    models.MoneyInCents
	ItemCategoryID  models.ID
	SellerID        models.ID
	Donation        bool
	Charity         bool
	Large           bool
}

type GetSoldItemsQuery struct{}

func NewGetSoldItemsQuery() *GetSoldItemsQuery {
	query := GetSoldItemsQuery{}

	return &query
}

func (q *GetSoldItemsQuery) Execute(db DatabaseQuerier) (r_result []*SoldItem, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	rows, err := db.Query(
		`
			SELECT
				sale.sale_id,
				sale.cashier_id,
				sale.transaction_time,
				item.item_id,
				item.added_at,
				item.description,
				item.price_in_cents,
				item.item_category_id,
				item.seller_id,
				item.donation,
				item.charity,
				item.large
			FROM
				sales sale
			INNER JOIN
				sale_items sale_item ON sale.sale_id = sale_item.sale_id
			INNER JOIN
				items item ON sale_item.item_id = item.item_id
			ORDER BY
				sale.sale_id, item.item_id
		`,
	)

	if err != nil {
		return nil, err
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	var soldItems []*SoldItem

	for rows.Next() {
		var saleID models.ID
		var cashierID models.ID
		var transactionTime models.Timestamp
		var itemID models.ID
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var itemCategory models.ID
		var sellerID models.ID
		var donation bool
		var charity bool
		var large bool

		err := rows.Scan(
			&saleID,
			&cashierID,
			&transactionTime,
			&itemID,
			&addedAt,
			&description,
			&priceInCents,
			&itemCategory,
			&sellerID,
			&donation,
			&charity,
			&large,
		)
		if err != nil {
			return nil, err
		}

		soldItem := SoldItem{
			SaleID:          saleID,
			CashierID:       cashierID,
			TransactionTime: transactionTime,
			ItemID:          itemID,
			AddedAt:         addedAt,
			Description:     description,
			PriceInCents:    priceInCents,
			ItemCategoryID:  itemCategory,
			SellerID:        sellerID,
			Donation:        donation,
			Charity:         charity,
			Large:           large,
		}

		soldItems = append(soldItems, &soldItem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return soldItems, nil
}
