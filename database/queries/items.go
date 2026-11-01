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
	"strings"
)

// GetItemIDs retrieves the IDs of all items in the database.
func GetItemIDs(ctx context.Context, db DatabaseQuerier) (r_result []models.ID, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	itemIDs, err := isql.New(db).GetItemIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query to look up item ids in database: %w", err)
	}

	return itemIDs, nil
}

// Returns the items associated with the given seller.
// The itemSelection parameter allows specifying whether to include visible/hidden items or not.
// The items are ordered by their time of addition, then by id.
// An ErrNoSuchUser is returned if no user with the given sellerID exists.
// An ErrWrongRole is returned if sellerID does not refer to a seller.
func GetSellerItems(ctx context.Context, db DatabaseQuerier, sellerID models.ID, itemSelection ItemSelection) (r_items []*models.Item, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	// Note: GetSellerItems performs multiple queries, but no transaction is necessary
	// since once a user exists with a certain role, it will not disappear.
	if err := EnsureUserExistsAndHasRole(ctx, db, sellerID, models.NewSellerRoleID()); err != nil {
		return nil, err
	}

	rows, err := isql.New(db).GetSellerItems(ctx, isql.GetSellerItemsParams{
		SellerID: sellerID.Int64(),
		Hidden:   itemSelectionToNullBool(itemSelection),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute query to get seller item data from database: %w", err)
	}

	items := make([]*models.Item, 0)

	for i := range rows {
		items = append(items, &models.Item{
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
		})
	}

	return items, nil
}

type ItemWithSaleCount struct {
	models.Item
	SaleCount int
}

// GetItemsWithSaleCounts looks up all items and how often they have been sold.
// Specifying a non-nil sellerID will return only items from that seller, otherwise items from all sellers are returned.
// The items are ordered by their time of addition, then by id.
// An ErrNoSuchUser is returned if no user with the given sellerID exists.
// An ErrWrongRole is returned if sellerID does not refer to a seller.
func GetItemsWithSaleCounts(ctx context.Context, db DatabaseQuerier, itemSelection ItemSelection, sellerID *models.ID) (r_items []*ItemWithSaleCount, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	// Note: GetSellerItems performs multiple queries, but no transaction is necessary
	// since once a user exists with a certain role, it will not disappear.
	if sellerID != nil {
		if err := EnsureUserExistsAndHasRole(ctx, db, *sellerID, models.NewSellerRoleID()); err != nil {
			return nil, err
		}
	}

	var sellerFilter sql.NullInt64
	if sellerID != nil {
		sellerFilter = sql.NullInt64{Int64: sellerID.Int64(), Valid: true}
	}

	rows, err := isql.New(db).GetItemsWithSaleCounts(ctx, isql.GetItemsWithSaleCountsParams{
		Hidden:   itemSelectionToNullBool(itemSelection),
		SellerID: sellerFilter,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute query to get seller items with sale counts from database: %w", err)
	}

	itemsWithSaleCount := make([]*ItemWithSaleCount, 0)

	for i := range rows {
		itemsWithSaleCount = append(itemsWithSaleCount, &ItemWithSaleCount{
			Item: models.Item{
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
			},
			SaleCount: int(rows[i].SaleCount),
		})
	}

	return itemsWithSaleCount, nil
}

// Returns the items associated with the given seller.
// The items are ordered by their time of addition, then by id.
// Hidden items are not included, as they cannot be sold.
// An ErrNoSuchUser is returned if no user with the given sellerID exists.
// An ErrWrongRole is returned if sellerID does not refer to a seller.
func GetSellerItemsWithSaleCounts(ctx context.Context, db DatabaseQuerier, sellerID models.ID) (r_items []*ItemWithSaleCount, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	// Note: GetSellerItems performs multiple queries, but no transaction is necessary
	// since once a user exists with a certain role, it will not disappear.
	if err := EnsureUserExistsAndHasRole(ctx, db, sellerID, models.NewSellerRoleID()); err != nil {
		return nil, err
	}

	rows, err := isql.New(db).GetSellerItemsWithSaleCounts(ctx, sellerID.Int64())
	if err != nil {
		return nil, fmt.Errorf("failed to execute query to get seller items with sale counts from database: %w", err)
	}

	itemsWithSaleCount := make([]*ItemWithSaleCount, 0)

	for i := range rows {
		itemsWithSaleCount = append(itemsWithSaleCount, &ItemWithSaleCount{
			Item: models.Item{
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
			},
			SaleCount: int(rows[i].SaleCount),
		})
	}

	return itemsWithSaleCount, nil
}

// Returns the item with the given identifier.
// A ErrNoSuchItem is returned if no item with the given identifier exists.
func GetItemWithID(ctx context.Context, db DatabaseQuerier, itemID models.ID) (r_result *models.Item, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	row, err := isql.New(db).GetItemWithID(ctx, itemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dberr.ErrNoSuchItem
		}
		return nil, fmt.Errorf("failed to read row: %w", err)
	}

	item := models.Item{
		ItemID:       itemID,
		AddedAt:      models.Timestamp(row.AddedAt),
		Description:  row.Description,
		PriceInCents: models.MoneyInCents(row.PriceInCents),
		CategoryID:   models.ID(row.ItemCategoryID),
		SellerID:     models.ID(row.SellerID),
		Donation:     row.Donation,
		Charity:      row.Charity,
		Frozen:       row.Frozen,
		Hidden:       row.Hidden,
		Large:        row.Large,
	}
	return &item, nil
}

// GetItemsWithIDs looks up items with the given IDs.
// The result is a map that relates item IDs to the corresponding item.
// Duplicates in itemIDs are ignored.
// If itemIDs contains a nonexistent item id, a ErrNoSuchItem is returned.
func GetItemsWithIDs(ctx context.Context, db DatabaseQuerier, itemIDs []models.ID) (r_result map[models.ID]*models.Item, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	items := make(map[models.ID]*models.Item)

	query := NewGetItemsQuery()
	query.WithItemIDs(itemIDs)

	query.Execute(ctx, db, func(item *models.Item) error {
		items[item.ItemID] = item
		return nil
	})

	// Check if all requested items were found
	if len(items) != len(itemIDs) {
		for _, itemID := range itemIDs {
			if _, ok := items[itemID]; !ok {
				return nil, fmt.Errorf("while getting items, among which %d: %w", itemID, dberr.ErrNoSuchItem)
			}
		}

		// If we get past the loop, it means that all items were found
		// There were duplicates in the requested IDs, but this is not an error
	}

	return items, nil
}

type ItemStatisticsResult struct {
	ItemCount         int
	TotalValueInCents models.MoneyInCents
}

type GetItemStatisticsQuery struct {
	HiddenFilter
	DescriptionFilter
	CategoryFilter
}

func NewGetItemStatisticsQuery() *GetItemStatisticsQuery {
	query := GetItemStatisticsQuery{
		HiddenFilter:      HiddenFilter{hidden: nil},
		DescriptionFilter: DescriptionFilter{descriptionPattern: nil},
		CategoryFilter:    CategoryFilter{categoryID: nil},
	}

	return &query
}

func (q *GetItemStatisticsQuery) Execute(ctx context.Context, db DatabaseQuerier) (r_result *ItemStatisticsResult, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	params := isql.GetItemStatisticsParams{
		Hidden:         sql.NullBool{},
		Description:    sql.NullString{},
		ItemCategoryID: sql.NullInt64{},
	}

	if q.hidden != nil {
		params.Hidden = sql.NullBool{Bool: *q.hidden, Valid: true}
	}

	if q.descriptionPattern != nil {
		params.Description = sql.NullString{String: *q.descriptionPattern, Valid: true}
	}

	if q.categoryID != nil {
		params.ItemCategoryID = sql.NullInt64{Int64: q.categoryID.Int64(), Valid: true}
	}

	row, err := isql.New(db).GetItemStatistics(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to read row: %w", err)
	}

	result := ItemStatisticsResult{
		ItemCount:         int(row.Count),
		TotalValueInCents: models.MoneyInCents(row.Column2),
	}
	return &result, nil
}

// ItemWithIDExists returns true if an item with the given identifier exists in the database.
func ItemWithIDExists(ctx context.Context, db DatabaseQuerier, itemID models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	_, err := isql.New(db).ItemWithIDExists(ctx, itemID)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

// ItemsExists checks if all given items exist in the database.
// Duplicates in itemIDs have no effect on the result.
// Returns true if all items exist, false otherwise.
func ItemsExist(ctx context.Context, db DatabaseQuerier, itemIDs []models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if len(itemIDs) == 0 {
		return true, nil
	}

	itemIDs = algorithms.RemoveDuplicates(itemIDs)

	count, err := isql.New(db).CountItemsWithIDs(ctx, itemIDs)
	if err != nil {
		return false, err
	}

	return int(count) == len(itemIDs), nil
}

// EnsureItemsExist checks if all items with the given IDs exist in the database.
// If any item does not exist, it returns a ErrNoSuchItem.
func EnsureItemsExist(ctx context.Context, db DatabaseQuerier, itemIDs []models.ID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	itemsExist, err := ItemsExist(ctx, db, itemIDs)
	if err != nil {
		return fmt.Errorf("failed to ensure items exist: %w", err)
	}
	if !itemsExist {
		return fmt.Errorf("failed to ensure items exist: %w", dberr.ErrNoSuchItem)
	}

	return nil
}

// UpdateFreezeStatusOfItems updates the frozen status of multiple items at once.
// If any item does not exist, it returns an ErrNoSuchItem.
// If any item is hidden, it returns a ErrItemHidden error.
// Duplicates in itemIDs are ignored.
// In case of an error, no items are updated.
// This function consists of multiple database interactions, so it must run within a transaction.
func UpdateFreezeStatusOfItems(ctx context.Context, transaction *TransactionalDatabaseQuerier, itemIDs []models.ID, frozen bool) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if len(itemIDs) == 0 {
		return nil
	}

	itemIDs = algorithms.RemoveDuplicates(itemIDs)

	if err := EnsureItemsExist(ctx, transaction, itemIDs); err != nil {
		return err
	}

	if err := EnsureNoHiddenItems(ctx, transaction, itemIDs); err != nil {
		return fmt.Errorf("failed to ensure no hidden items: %w", err)
	}

	if err := isql.New(transaction).UpdateFreezeStatusOfItems(ctx, isql.UpdateFreezeStatusOfItemsParams{
		Frozen:  frozen,
		ItemIds: itemIDs,
	}); err != nil {
		return err
	}

	return nil
}

// UpdateHiddenStatusOfItems updates the hidden status of multiple items at once.
// If any item does not exist, it returns an ErrNoSuchItem.
// If any item is hidden, it returns a ErrItemFrozen error.
// Duplicates in itemIDs are ignored.
// In case of an error, no items are updated.
// This function consists of multiple database interactions, so it must run within a transaction.
func UpdateHiddenStatusOfItems(ctx context.Context, transaction *TransactionalDatabaseQuerier, itemIDs []models.ID, hidden bool) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if len(itemIDs) == 0 {
		return nil
	}

	itemIDs = algorithms.RemoveDuplicates(itemIDs)

	if err := EnsureItemsExist(ctx, transaction, itemIDs); err != nil {
		return err
	}

	// Check if none of the items are frozen
	if err := EnsureNoFrozenItems(ctx, transaction, itemIDs); err != nil {
		return err
	}

	if err := isql.New(transaction).UpdateHiddenStatusOfItems(ctx, isql.UpdateHiddenStatusOfItemsParams{
		Hidden:  hidden,
		ItemIds: itemIDs,
	}); err != nil {
		return err
	}

	return nil
}

// PartitionItemsByHiddenStatus partitions the given item IDs into two sets: one for unhidden items and one for hidden items.
// If an item ID does not exist in the database, it is ignored.
func PartitionItemsByHiddenStatus(ctx context.Context, db DatabaseQuerier, itemIDs []models.ID) (r_visible *algorithms.Set[models.ID], r_hidden *algorithms.Set[models.ID], r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	rows, err := isql.New(db).GetHiddenStatusOfItems(ctx, itemIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query items: %w", err)
	}

	visible := algorithms.NewSet[models.ID]()
	hidden := algorithms.NewSet[models.ID]()

	for i := range rows {
		if rows[i].Hidden {
			hidden.Add(rows[i].ItemID)
		} else {
			visible.Add(rows[i].ItemID)
		}
	}

	return &visible, &hidden, nil
}

// PartitionItemsByFrozenStatus partitions the given item IDs into two sets: one for nonfrozen items and one for frozen items.
// If an item ID does not exist in the database, it is ignored.
func PartitionItemsByFrozenStatus(ctx context.Context, db DatabaseQuerier, itemIDs []models.ID) (r_nonfrozen *algorithms.Set[models.ID], r_frozen *algorithms.Set[models.ID], r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	rows, err := isql.New(db).GetFrozenStatusOfItems(ctx, itemIDs)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query items: %w", err)
	}

	nonfrozen := algorithms.NewSet[models.ID]()
	frozen := algorithms.NewSet[models.ID]()

	for i := range rows {
		if rows[i].Frozen {
			frozen.Add(rows[i].ItemID)
		} else {
			nonfrozen.Add(rows[i].ItemID)
		}
	}

	return &nonfrozen, &frozen, nil
}

// ContainsHiddenItems checks if any of the given items are hidden.
// It returns true if at least one item is hidden, and false otherwise.
// It is not an error when nonexistent items are passed in, they are simply ignored.
func ContainsHiddenItems(ctx context.Context, database DatabaseQuerier, itemIDs []models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if len(itemIDs) == 0 {
		return false, nil
	}

	_, hidden, err := PartitionItemsByHiddenStatus(ctx, database, itemIDs)
	if err != nil {
		return false, err
	}

	containsHidden := hidden.Len() > 0

	return containsHidden, nil
}

// ContainsFrozenItems checks if any of the given items are frozen.
// It returns true if at least one item is frozen, and false otherwise.
// It is not an error when nonexistent items are passed in, they are simply ignored.
func ContainsFrozenItems(ctx context.Context, database DatabaseQuerier, itemIDs []models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if len(itemIDs) == 0 {
		return false, nil
	}

	_, frozen, err := PartitionItemsByFrozenStatus(ctx, database, itemIDs)
	if err != nil {
		return false, err
	}

	containsFrozen := frozen.Len() > 0

	return containsFrozen, nil
}

// IsItemFrozen checks if none of the items is frozen.
// If one or more items are frozen, it returns an ErrItemFrozen error.
func EnsureNoFrozenItems(ctx context.Context, qh DatabaseQuerier, itemIDs []models.ID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	containsFrozen, err := ContainsFrozenItems(ctx, qh, itemIDs)

	if err != nil {
		return fmt.Errorf("failed to check for frozen items: %w", err)
	}

	if containsFrozen {
		return dberr.ErrItemFrozen
	}

	return nil
}

// EnsureNoHiddenItems checks if none of the items is hidden.
// If one or more items are hidden, it returns an ErrItemHidden error.
func EnsureNoHiddenItems(ctx context.Context, qh DatabaseQuerier, itemIDs []models.ID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	containsHidden, err := ContainsHiddenItems(ctx, qh, itemIDs)

	if err != nil {
		return fmt.Errorf("failed to check for hidden items: %w", err)
	}

	if containsHidden {
		return dberr.ErrItemHidden
	}

	return nil
}

// IsItemFrozen checks whether the item with the given ID is frozen.
// ErrNoSuchItem is returned if the item does not exist.
func IsItemFrozen(ctx context.Context, db DatabaseQuerier, itemID models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	nonFrozen, frozen, err := PartitionItemsByFrozenStatus(ctx, db, []models.ID{itemID})
	if err != nil {
		return false, err
	}

	isFrozen := frozen.Len() > 0
	if isFrozen {
		return true, nil
	}

	isUnfrozen := nonFrozen.Len() > 0
	if isUnfrozen {
		return false, nil
	}

	return false, fmt.Errorf("failed to check if item %d is frozen: %w", itemID, dberr.ErrNoSuchItem)
}

// IsItemHidden checks whether the item with the given ID is hidden.
// ErrNoSuchItem is returned if the item does not exist.
func IsItemHidden(ctx context.Context, db DatabaseQuerier, itemID models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	unhidden, hidden, err := PartitionItemsByHiddenStatus(ctx, db, []models.ID{itemID})
	if err != nil {
		return false, err
	}

	isFrozen := hidden.Len() > 0
	if isFrozen {
		return true, nil
	}

	isUnfrozen := unhidden.Len() > 0
	if isUnfrozen {
		return false, nil
	}

	return false, fmt.Errorf("failed to check if item %d is hidden: %w", itemID, dberr.ErrNoSuchItem)
}

// RemoveItemWithID removes the item with the given ID from the database.
// This function breaks the monotonicity invariant upon which other functionality relies,
// so it should be used with caution.
// If the item does not exist, ErrNoSuchItem is returned.
// If the item has been sold, ErrItemSold is returned and the item remains in the database.
// If the item is frozen, ErrItemFrozen is returned and the item remains in the database.
func RemoveItemWithID(ctx context.Context, db DatabaseQuerier, itemID models.ID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if err := EnsureItemsExist(ctx, db, []models.ID{itemID}); err != nil {
		return fmt.Errorf("failed to remove item with id %d: %w", itemID, dberr.ErrNoSuchItem)
	}

	if err := EnsureNoFrozenItems(ctx, db, []models.ID{itemID}); err != nil {
		return fmt.Errorf("failed to remove item with id %d: %w", itemID, dberr.ErrItemFrozen)
	}

	if err := isql.New(db).RemoveItemWithID(ctx, itemID); err != nil {
		sold, err2 := HasAnyBeenSold(ctx, db, []models.ID{itemID})
		if err2 != nil {
			return err
		}
		if sold {
			return fmt.Errorf("failed to remove item with id %d: %w", itemID, dberr.ErrItemSold)
		}

		return fmt.Errorf("failed to remove item with id %d: %w", itemID, err)
	}

	return nil
}

// ItemUpdate represents the fields that can be updated in an item.
// It is used by the function UpdateItem.
type ItemUpdate struct {
	AddedAt      *models.Timestamp    // If nil, the AddedAt field is not updated.
	Description  *string              // If nil, the Description field is not updated.
	PriceInCents *models.MoneyInCents // If nil, the PriceInCents field is not updated.
	CategoryID   *models.ID           // If nil, the CategoryID field is not updated.
	Donation     *bool                // If nil, the Donation field is not updated.
	Charity      *bool                // If nil, the Charity field is not updated.
	Large        *bool                // If nil, the Large field is not updated
}

// UpdateItem updates the item with the given ID in the database.
// If the item does not exist, ErrNoSuchItem is returned.
// If the item is frozen, ErrItemFrozen is returned.
// If the item is hidden, ErrItemHidden is returned.
func UpdateItem(ctx context.Context, db *TransactionalDatabaseQuerier, itemID models.ID, itemUpdate *ItemUpdate) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	item, err := GetItemWithID(ctx, db, itemID)
	if err != nil {
		return err
	}

	if item.Frozen {
		return dberr.ErrItemFrozen
	}

	if item.Hidden {
		return dberr.ErrItemHidden
	}

	sqlValues := isql.UpdateItemParams{ItemID: itemID}

	if itemUpdate.AddedAt != nil {
		sqlValues.AddedAt = sql.NullInt64{Int64: int64(*itemUpdate.AddedAt), Valid: true}
	}

	if itemUpdate.Description != nil {
		sqlValues.Description = sql.NullString{String: *itemUpdate.Description, Valid: true}
	}

	if itemUpdate.PriceInCents != nil {
		if !models.IsValidPrice(*itemUpdate.PriceInCents) {
			return fmt.Errorf("failed to updated item's price to %d: %w", *itemUpdate.PriceInCents, dberr.ErrInvalidPrice)
		}

		sqlValues.PriceInCents = sql.NullInt64{Int64: int64(*itemUpdate.PriceInCents), Valid: true}
	}

	if itemUpdate.CategoryID != nil {
		categoryExists, err := CategoryWithIDExists(ctx, db, *itemUpdate.CategoryID)
		if err != nil {
			return err
		}

		if !categoryExists {
			return fmt.Errorf("failed to update item's category to %d", *itemUpdate.CategoryID)
		}

		sqlValues.ItemCategoryID = sql.NullInt64{Int64: itemUpdate.CategoryID.Int64(), Valid: true}
	}

	if itemUpdate.Donation != nil {
		sqlValues.Donation = sql.NullBool{Bool: *itemUpdate.Donation, Valid: true}
	}

	if itemUpdate.Charity != nil {
		sqlValues.Charity = sql.NullBool{Bool: *itemUpdate.Charity, Valid: true}
	}

	if itemUpdate.Large != nil {
		sqlValues.Large = sql.NullBool{Bool: *itemUpdate.Large, Valid: true}
	}

	if !sqlValues.AddedAt.Valid && !sqlValues.Description.Valid && !sqlValues.PriceInCents.Valid && !sqlValues.ItemCategoryID.Valid && !sqlValues.Donation.Valid && !sqlValues.Charity.Valid && !sqlValues.Large.Valid {
		return nil
	}

	if err := isql.New(db).UpdateItem(ctx, sqlValues); err != nil {
		return err
	}

	return nil
}

type AddItemFunction func(addedAt models.Timestamp, description string, priceInCents models.MoneyInCents, itemCategoryID models.ID, sellerID models.ID, donation bool, charity bool, frozen bool, hidden bool, large bool)

type AddItemsCallback func(addItem AddItemFunction)

// NOTE: AddItems is kept as raw SQL: sqlc's SQLite engine cannot capture a
// query parameter passed to the json_each table-valued function, which the
// batch insert would require. See docs/superpowers/specs/2026-11-01-sqlc-query-migration-design.md.
func AddItems(ctx context.Context, db DatabaseQuerier, callback AddItemsCallback) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	valuesString := []string{}
	arguments := []any{}
	tupleString := "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

	add := func(addedAt models.Timestamp, description string, priceInCents models.MoneyInCents, itemCategoryID models.ID, sellerID models.ID, donation bool, charity bool, frozen bool, hidden bool, large bool) {
		valuesString = append(valuesString, tupleString)
		arguments = append(arguments, addedAt, description, priceInCents, itemCategoryID, sellerID, donation, charity, frozen, hidden, large)
	}

	callback(add)

	if len(valuesString) == 0 {
		return nil
	}

	query := `INSERT INTO items (added_at, description, price_in_cents, item_category_id, seller_id, donation, charity, frozen, hidden, large) VALUES ` + strings.Join(valuesString, ",")

	if _, err := db.ExecContext(ctx, query, arguments...); err != nil {
		return err
	}

	return nil
}

// DoesSellerHaveFrozenItems checks if any item owned by the given seller is frozen.
// Returns ErrNoSuchUser is sellerID does not exist.
// Returns ErrWrongRole if sellerID does not refer to a seller.
func DoesSellerHaveFrozenItems(ctx context.Context, db *TransactionalDatabaseQuerier, sellerID models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if err := EnsureUserExistsAndHasRole(ctx, db, sellerID, models.NewSellerRoleID()); err != nil {
		return false, err
	}

	_, err := isql.New(db).DoesSellerHaveFrozenItems(ctx, sellerID.Int64())

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// MoveItemsToNewSeller moves all items belonging to oldSellerID to newSellerID.
// Does not check for frozen or hidden items!
// Returns ErrNoSuchUser if oldSellerID or newSellerID do not exist.
// Returns ErrWrongRole if oldSellerID or newSellerID do not refer to sellers.
func MoveItemsToNewSeller(ctx context.Context, db *TransactionalDatabaseQuerier, oldSellerID models.ID, newSellerID models.ID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	// oldSellerID must refer to seller
	if err := EnsureUserExistsAndHasRole(ctx, db, oldSellerID, models.NewSellerRoleID()); err != nil {
		return err
	}

	// newSellerID must refer to seller
	if err := EnsureUserExistsAndHasRole(ctx, db, newSellerID, models.NewSellerRoleID()); err != nil {
		return err
	}

	if err := isql.New(db).MoveItemsToNewSeller(ctx, isql.MoveItemsToNewSellerParams{
		SellerID:   newSellerID.Int64(),
		SellerID_2: oldSellerID.Int64(),
	}); err != nil {
		return err
	}

	return nil
}
