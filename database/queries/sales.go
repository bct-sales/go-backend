package queries

import (
	"bctbackend/algorithms"
	"context"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries/internal/isql"
	"database/sql"
	"errors"
	"fmt"
)

type addSaleQuery struct {
	CashierID       models.ID
	TransactionTime models.Timestamp
	ItemIDs         []models.ID
}

func (q *addSaleQuery) execute(ctx context.Context, db *TransactionalDatabaseQuerier) (r_result models.ID, r_err error) {
	if err := q.ensureInputsValidity(ctx, db); err != nil {
		return 0, err
	}

	// Check if all items exist
	if err := EnsureItemsExist(ctx, db, q.ItemIDs); err != nil {
		return 0, fmt.Errorf("failed to add sale: %w", dberr.ErrNoSuchItem)
	}

	// Check if any of the items are hidden
	if err := EnsureNoHiddenItems(ctx, db, q.ItemIDs); err != nil {
		return 0, err
	}

	// Create sale
	saleID, err := isql.New(db).AddSale(ctx, isql.AddSaleParams{
		CashierID:       q.CashierID,
		TransactionTime: int64(q.TransactionTime),
	})
	if err != nil {
		return 0, err
	}

	// Add items to sale
	for _, itemID := range q.ItemIDs {
		err := isql.New(db).AddSaleItem(ctx, isql.AddSaleItemParams{
			SaleID: models.ID(saleID),
			ItemID: itemID,
		})

		if err != nil {
			return 0, err
		}
	}

	return models.ID(saleID), nil
}

func (q *addSaleQuery) ensureInputsValidity(ctx context.Context, db DatabaseQuerier) error {
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
	cashier, err := GetUserWithID(ctx, db, q.CashierID)
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
	ctx context.Context,
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

	return query.execute(ctx, db)
}

type GetSalesQuery struct {
	minimalID         *models.ID // If set, only sales with an ID greater than or equal to this value are returned.
	antiChronological bool       // Specifies whether the results are returned in anti-chronological order
	rowSelection      RowRange
}

func NewGetSalesQuery() *GetSalesQuery {
	return &GetSalesQuery{
		minimalID:         nil,
		rowSelection:      RowRange{Limit: nil, Offset: nil},
		antiChronological: false,
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
	q.antiChronological = true
	return q
}

func (q *GetSalesQuery) Execute(ctx context.Context, db DatabaseQuerier, receiver func(*models.SaleSummary) error) (r_err error) {
	limit, offset := q.rowSelection.LimitOffset()
	params := isql.GetSalesParams{
		MinimalID: sql.NullInt64{},
		Limit:     limit,
		Offset:    offset,
	}

	if q.minimalID != nil {
		params.MinimalID = sql.NullInt64{Int64: q.minimalID.Int64(), Valid: true}
	}

	toSummary := func(row isql.GetSalesRow) *models.SaleSummary {
		return &models.SaleSummary{
			SaleID:            models.ID(row.SaleID),
			CashierID:         row.CashierID,
			TransactionTime:   models.Timestamp(row.TransactionTime),
			ItemCount:         int(row.ItemCount),
			TotalPriceInCents: models.MoneyInCents(row.TotalPrice),
		}
	}

	if !q.antiChronological {
		rows, err := isql.New(db).GetSales(ctx, params)
		if err != nil {
			return err
		}

		for i := range rows {
			if err := receiver(toSummary(rows[i])); err != nil {
				return err
			}
		}
	} else {
		rows, err := isql.New(db).GetSalesAntiChronologically(ctx, isql.GetSalesAntiChronologicallyParams(params))
		if err != nil {
			return err
		}

		for i := range rows {
			if err := receiver(toSummary(isql.GetSalesRow(rows[i]))); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetSaleWithID returns the sale with the given saleID.
// A ErrNoSuchSale is returned if no sale with the given saleID exists.
func GetSaleWithID(ctx context.Context, db DatabaseQuerier, saleID models.ID) (r_result *models.Sale, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	row, err := isql.New(db).GetSaleWithID(ctx, saleID.Int64())
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to get sale with id %d: %w", saleID, dberr.ErrNoSuchSale)
	}
	if err != nil {
		return nil, err
	}

	sale := models.Sale{
		SaleID:          saleID,
		CashierID:       models.ID(row.CashierID),
		TransactionTime: models.Timestamp(row.TransactionTime),
	}
	return &sale, nil
}

// SaleWithIDExists checks whether a sale with the given saleID exists in the database.
// Returns true if the sale exists, false otherwise.
func SaleWithIDExists(ctx context.Context, db DatabaseQuerier, saleID models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	_, err := isql.New(db).SaleWithIDExists(ctx, saleID.Int64())

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
func GetSaleItems(ctx context.Context, db DatabaseQuerier, saleID models.ID) (r_result []*models.Item, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	saleExists, err := SaleWithIDExists(ctx, db, saleID)
	if err != nil {
		return nil, err
	}
	if !saleExists {
		return nil, fmt.Errorf("failed to get items of sale %d: %w", saleID, dberr.ErrNoSuchSale)
	}

	rows, err := isql.New(db).GetSaleItems(ctx, saleID)
	if err != nil {
		return nil, err
	}

	var items []*models.Item
	for i := range rows {
		item := models.Item{
			ItemID:       rows[i].ItemID,
			AddedAt:      models.Timestamp(rows[i].AddedAt),
			Description:  rows[i].Description,
			PriceInCents: models.MoneyInCents(rows[i].PriceInCents),
			CategoryID:   models.ID(rows[i].ItemCategoryID),
			SellerID:     models.ID(rows[i].SellerID),
			Donation:     rows[i].Donation,
			Charity:      rows[i].Charity,
			Frozen:       rows[i].Frozen,
			Hidden:       false,
			Large:        rows[i].Large,
		}
		items = append(items, &item)
	}

	return items, nil
}

// RemoveSale removes the sale with the given saleID from the database.
// Returns ErrNoSuchSale if no such sale exists.
// This function should be use with care.
func RemoveSale(ctx context.Context, transaction *TransactionalDatabaseQuerier, saleID models.ID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	saleExists, err := SaleWithIDExists(ctx, transaction, saleID)
	if err != nil {
		return err
	}
	if !saleExists {
		return fmt.Errorf("failed to remove sale %d: %w", saleID, dberr.ErrNoSuchSale)
	}

	if err := isql.New(transaction).RemoveSaleItems(ctx, saleID); err != nil {
		return err
	}

	if err := isql.New(transaction).RemoveSale(ctx, saleID.Int64()); err != nil {
		return err
	}

	return nil
}

// GetSoldItems returns a list of all items that have been sold.
// The items are ordered by transaction time (most recent first) and item ID (lowest first).
func GetSoldItems(ctx context.Context, db DatabaseQuerier) (r_result []*models.Item, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	rows, err := isql.New(db).GetSoldItems(ctx)
	if err != nil {
		return nil, err
	}

	var items []*models.Item
	for i := range rows {
		item := models.Item{
			ItemID:       rows[i].ItemID,
			AddedAt:      models.Timestamp(rows[i].AddedAt),
			Description:  rows[i].Description,
			PriceInCents: models.MoneyInCents(rows[i].PriceInCents),
			CategoryID:   models.ID(rows[i].ItemCategoryID),
			SellerID:     models.ID(rows[i].SellerID),
			Donation:     rows[i].Donation,
			Charity:      rows[i].Charity,
			Frozen:       rows[i].Frozen,
			Hidden:       false,
			Large:        rows[i].Large,
		}
		items = append(items, &item)
	}

	return items, nil
}

// CountSoldItems returns the total number of items that have been sold.
// Two counts are returned: one where each item is counted only once, and one where each item is counted for each sale it was involved in.
func CountSoldItems(ctx context.Context, db DatabaseQuerier) (r_result *struct {
	Distinct         int
	IncludeMultiples int
}, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	row, err := isql.New(db).CountSoldItems(ctx)
	if err != nil {
		return nil, err
	}

	result := struct {
		Distinct         int
		IncludeMultiples int
	}{
		Distinct:         int(row.Count_2),
		IncludeMultiples: int(row.Count),
	}
	return &result, nil
}

// HasAnyBeenSold checks if any one of the given item was involved in one or more sales.
// Does not check if items exist.
func HasAnyBeenSold(ctx context.Context, db DatabaseQuerier, itemIDs []models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	count, err := isql.New(db).HasAnyBeenSold(ctx, itemIDs)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetItemsSoldBy returns a list of all items sold by a specified cashier.
// The items are ordered by transaction time (most recent first) and item ID (lowest first).
func GetItemsSoldBy(ctx context.Context, db DatabaseQuerier, cashierID models.ID) (r_result []*models.Item, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if err := EnsureUserExistsAndHasRole(ctx, db, cashierID, models.NewCashierRoleID()); err != nil {
		return nil, err
	}

	rows, err := isql.New(db).GetItemsSoldBy(ctx, cashierID)
	if err != nil {
		return nil, err
	}

	var items []*models.Item
	for i := range rows {
		item := models.Item{
			ItemID:       rows[i].ItemID,
			AddedAt:      models.Timestamp(rows[i].AddedAt),
			Description:  rows[i].Description,
			PriceInCents: models.MoneyInCents(rows[i].PriceInCents),
			CategoryID:   models.ID(rows[i].ItemCategoryID),
			SellerID:     models.ID(rows[i].SellerID),
			Donation:     rows[i].Donation,
			Charity:      rows[i].Charity,
			Frozen:       rows[i].Frozen,
			Hidden:       rows[i].Hidden,
			Large:        rows[i].Large,
		}
		items = append(items, &item)
	}

	return items, nil
}

// GetSalesWithItem returns a list of the ids of all sales that include a specified item.
// The ids are returned in ascending order.
func GetSalesWithItem(ctx context.Context, db DatabaseQuerier, itemID models.ID) (r_result []models.ID, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if itemExists, err := ItemWithIDExists(ctx, db, itemID); err != nil || !itemExists {
		if !itemExists {
			return nil, fmt.Errorf("failed to get sales with item %d: %w", itemID, dberr.ErrNoSuchItem)
		}

		return nil, err
	}

	rows, err := isql.New(db).GetSalesWithItem(ctx, itemID)
	if err != nil {
		return nil, err
	}

	saleIDs := []models.ID{}
	saleIDs = append(saleIDs, rows...)

	return saleIDs, nil
}

// GetSalesWithCashier returns a list of all sales made by a specified cashier.
// The sales are ordered by transaction time (chronologically) and sale ID (lowest first).
// Returns ErrNoSuchUser if the cashierID does not correspond to any user.
// Returns ErrWrongRole if the cashierID does not correspond to a cashier.
func GetSalesWithCashier(ctx context.Context, db DatabaseQuerier, cashierID models.ID) (r_result []*models.Sale, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if err := EnsureUserExistsAndHasRole(ctx, db, cashierID, models.NewCashierRoleID()); err != nil {
		return nil, err
	}

	rows, err := isql.New(db).GetSalesWithCashier(ctx, cashierID)
	if err != nil {
		return nil, err
	}

	sales := []*models.Sale{}
	for i := range rows {
		sale := models.Sale{
			SaleID:          models.ID(rows[i].SaleID),
			CashierID:       models.ID(rows[i].CashierID),
			TransactionTime: models.Timestamp(rows[i].TransactionTime),
		}
		sales = append(sales, &sale)
	}

	return sales, nil
}

// RemoveAllSales removes all sales from the database.
func RemoveAllSales(ctx context.Context, transaction *TransactionalDatabaseQuerier) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if err := isql.New(transaction).RemoveAllSaleItems(ctx); err != nil {
		return err
	}

	if err := isql.New(transaction).RemoveAllSales(ctx); err != nil {
		return err
	}

	return nil
}

// GetCashierSales retrieves a list of sales made by a specified cashier.
// If cashierID does not correspond to any user, ErrNoSuchUser is returned.
// If cashierID does not correspond to a cashier, ErrWrongRole is returned.
func GetCashierSales(ctx context.Context, db DatabaseQuerier, cashierID models.ID, receiver func(*models.SaleSummary) error, order Order, rowSelection *RowRange) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if err := EnsureUserExistsAndHasRole(ctx, db, cashierID, models.NewCashierRoleID()); err != nil {
		return err
	}

	limit, offset := rowSelection.LimitOffset()
	params := isql.GetCashierSalesChronologicalParams{
		CashierID: cashierID,
		Limit:     limit,
		Offset:    offset,
	}

	toSummary := func(row isql.GetCashierSalesChronologicalRow) *models.SaleSummary {
		return &models.SaleSummary{
			SaleID:            models.ID(row.SaleID),
			CashierID:         row.CashierID,
			TransactionTime:   models.Timestamp(row.TransactionTime),
			ItemCount:         int(row.ItemCount),
			TotalPriceInCents: models.MoneyInCents(row.TotalPrice),
		}
	}

	switch order {
	case OrderChronological:
		rows, err := isql.New(db).GetCashierSalesChronological(ctx, params)
		if err != nil {
			return err
		}

		for i := range rows {
			if err := receiver(toSummary(rows[i])); err != nil {
				return err
			}
		}
	case OrderAntiChronological:
		rows, err := isql.New(db).GetCashierSalesAntiChronological(ctx, isql.GetCashierSalesAntiChronologicalParams(params))
		if err != nil {
			return err
		}

		for i := range rows {
			if err := receiver(toSummary(isql.GetCashierSalesChronologicalRow(rows[i]))); err != nil {
				return err
			}
		}
	}

	return nil
}

// CountSales returns the total number of sales in the database.
func CountSales(ctx context.Context, db DatabaseQuerier) (r_result int, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	count, err := isql.New(db).CountSales(ctx)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// CountCashierSales returns the total number of sales made by the specified cashier.
func CountCashierSales(ctx context.Context, db DatabaseQuerier, cashierID models.ID) (r_result int, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if err := EnsureUserExistsAndHasRole(ctx, db, cashierID, models.NewCashierRoleID()); err != nil {
		return 0, err
	}

	count, err := isql.New(db).CountCashierSales(ctx, cashierID)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// GetTotalSalesValue returns the total value of all sales in the database.
// The value is calculated as the sum of the prices of all items sold.
// If an item was sold multiple times, its price is counted each time.
func GetTotalSalesValue(ctx context.Context, db DatabaseQuerier) (r_result models.MoneyInCents, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	totalValue, err := isql.New(db).GetTotalSalesValue(ctx)
	if err != nil {
		return 0, err
	}

	return models.MoneyInCents(totalValue), nil
}

type MultiplySoldItem struct {
	Item  models.Item
	Sales []models.Sale
}

// GetMultiplySoldItems retrieves all items that have been sold multiple times.
func GetMultiplySoldItems(ctx context.Context, db DatabaseQuerier) (r_result []MultiplySoldItem, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	rows, err := isql.New(db).GetMultiplySoldItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	var multiplySoldItems []MultiplySoldItem

	for i := range rows {
		rowData := rows[i]

		sale := models.Sale{
			SaleID:          models.ID(rowData.SaleID),
			CashierID:       rowData.CashierID,
			TransactionTime: models.Timestamp(rowData.TransactionTime),
		}

		lastMultiplySoldItemIndex := len(multiplySoldItems) - 1

		if lastMultiplySoldItemIndex >= 0 && multiplySoldItems[lastMultiplySoldItemIndex].Item.ItemID == rowData.ItemID {
			multiplySoldItems[lastMultiplySoldItemIndex].Sales = append(multiplySoldItems[lastMultiplySoldItemIndex].Sales, sale)
		} else {
			multiplySoldItem := MultiplySoldItem{
				Item: models.Item{
					ItemID:       rowData.ItemID,
					AddedAt:      models.Timestamp(rowData.AddedAt),
					Description:  rowData.Description,
					PriceInCents: models.MoneyInCents(rowData.PriceInCents),
					CategoryID:   models.ID(rowData.ItemCategoryID),
					SellerID:     models.ID(rowData.SellerID),
					Donation:     rowData.Donation,
					Charity:      rowData.Charity,
					Frozen:       rowData.Frozen,
					Hidden:       false,
					Large:        rowData.Large,
				},
				Sales: []models.Sale{sale},
			}

			multiplySoldItems = append(multiplySoldItems, multiplySoldItem)
		}
	}

	return multiplySoldItems, nil
}

type SaleItemInformation struct {
	SellerID       models.ID
	Description    string
	ItemCategoryID models.ID
	PriceInCents   models.MoneyInCents
	Large          bool
	SellCount      int64
}

// GetSaleItemInformation retrieves information about a sale item.
// If the item is not found, it returns an ErrNoSuchItem.
func GetSaleItemInformation(
	ctx context.Context,
	db DatabaseQuerier,
	itemID models.ID) (r_result *SaleItemInformation, r_err error) {

	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	row, err := isql.New(db).GetSaleItemInformation(ctx, itemID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to get information about item %d: %w", itemID, dberr.ErrNoSuchItem)
	}
	if err != nil {
		return nil, err
	}

	saleItemInformation := SaleItemInformation{
		SellerID:       models.ID(row.SellerID),
		Description:    row.Description,
		ItemCategoryID: models.ID(row.ItemCategoryID),
		PriceInCents:   models.MoneyInCents(row.PriceInCents),
		Large:          row.Large,
		SellCount:      row.Count,
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
func GetSalesOverview(ctx context.Context, db DatabaseQuerier) (r_result []CategorySaleTotal, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	rows, err := isql.New(db).GetSalesOverview(ctx)
	if err != nil {
		return nil, err
	}

	var categorySaleTotals []CategorySaleTotal

	for i := range rows {
		categorySaleTotal := CategorySaleTotal{
			CategoryID:   rows[i].ItemCategoryID,
			CategoryName: rows[i].Name,
			TotalInCents: models.MoneyInCents(rows[i].Total),
		}

		categorySaleTotals = append(categorySaleTotals, categorySaleTotal)
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

func (q *GetSoldItemsQuery) Execute(ctx context.Context, db DatabaseQuerier) (r_result []*SoldItem, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	rows, err := isql.New(db).ListSoldItems(ctx)
	if err != nil {
		return nil, err
	}

	var soldItems []*SoldItem

	for i := range rows {
		soldItem := SoldItem{
			SaleID:          models.ID(rows[i].SaleID),
			CashierID:       rows[i].CashierID,
			TransactionTime: models.Timestamp(rows[i].TransactionTime),
			ItemID:          rows[i].ItemID,
			AddedAt:         models.Timestamp(rows[i].AddedAt),
			Description:     rows[i].Description,
			PriceInCents:    models.MoneyInCents(rows[i].PriceInCents),
			ItemCategoryID:  models.ID(rows[i].ItemCategoryID),
			SellerID:        models.ID(rows[i].SellerID),
			Donation:        rows[i].Donation,
			Charity:         rows[i].Charity,
			Large:           rows[i].Large,
		}

		soldItems = append(soldItems, &soldItem)
	}

	return soldItems, nil
}
