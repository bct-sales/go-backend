package queries

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func GetItems(db *sql.DB, receiver func(*models.Item) error, itemSelection ItemSelection) error {
	// Build SQL query
	query := fmt.Sprintf(`
		SELECT item_id, added_at, description, price_in_cents, item_category_id, seller_id, donation, charity, frozen, hidden
		FROM %s
		ORDER BY item_id ASC
	`, ItemsTableFor(itemSelection))

	// Perform query
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to execute query to look up items in database: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()

	// Iterate over rows and call receiver function for each item
	for rows.Next() {
		var id models.Id
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var itemCategoryId models.Id
		var sellerId models.Id
		var donation bool
		var charity bool
		var frozen bool
		var hidden bool
		err = rows.Scan(&id, &addedAt, &description, &priceInCents, &itemCategoryId, &sellerId, &donation, &charity, &frozen, &hidden)
		if err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		item := models.Item{
			ItemID:       id,
			AddedAt:      addedAt,
			Description:  description,
			PriceInCents: priceInCents,
			CategoryID:   itemCategoryId,
			SellerID:     sellerId,
			Donation:     donation,
			Charity:      charity,
			Frozen:       frozen,
			Hidden:       hidden,
		}

		if err := receiver(&item); err != nil {
			return fmt.Errorf("receiver failed: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return nil
}

// Returns the items associated with the given seller.
// The items are ordered by their time of addition, then by id.
// An NoSuchUserError is returned if no user with the given sellerId exists.
// An InvalidRoleError is returned if sellerId does not refer to a seller.
func GetSellerItems(db *sql.DB, sellerId models.Id, itemSelection ItemSelection) (r_items []*models.Item, r_err error) {
	if err := EnsureUserExistsAndHasRole(db, sellerId, models.NewSellerRoleId()); err != nil {
		return nil, err
	}

	// Build SQL query
	query := fmt.Sprintf(`
		SELECT item_id, added_at, description, price_in_cents, item_category_id, seller_id, donation, charity, frozen, hidden
		FROM %s
		WHERE seller_id = ?
		ORDER BY added_at, item_id ASC
	`, ItemsTableFor(itemSelection))

	rows, err := db.Query(query, sellerId)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query to get seller item data from database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	items := make([]*models.Item, 0)

	for rows.Next() {
		var id models.Id
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var itemCategoryId models.Id
		var sellerId models.Id
		var donation bool
		var charity bool
		var frozen bool
		var hidden bool

		err = rows.Scan(&id, &addedAt, &description, &priceInCents, &itemCategoryId, &sellerId, &donation, &charity, &frozen, &hidden)
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		item := models.Item{
			ItemID:       id,
			AddedAt:      addedAt,
			Description:  description,
			PriceInCents: priceInCents,
			CategoryID:   itemCategoryId,
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

type ItemWithSaleCount struct {
	models.Item
	SaleCount int
}

// Returns the items associated with the given seller.
// The items are ordered by their time of addition, then by id.
// Hidden items are not included, as they cannot be sold.
// An NoSuchUserError is returned if no user with the given sellerId exists.
// An InvalidRoleError is returned if sellerId does not refer to a seller.
func GetSellerItemsWithSaleCounts(db *sql.DB, sellerId models.Id) (r_items []*ItemWithSaleCount, r_err error) {
	if err := EnsureUserExistsAndHasRole(db, sellerId, models.NewSellerRoleId()); err != nil {
		return nil, err
	}

	rows, err := db.Query(
		`
			SELECT items.item_id, added_at, description, price_in_cents, item_category_id, seller_id, donation, charity, frozen, hidden, COALESCE(COUNT(sale_items.sale_id), 0) AS sale_count
			FROM items LEFT JOIN sale_items ON items.item_id = sale_items.item_id
			WHERE seller_id = ? AND hidden = false
			GROUP BY items.item_id
			ORDER BY added_at, items.item_id ASC
		`,
		sellerId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query to get seller items with sale counts from database: %w", err)
	}

	defer func() { err = errors.Join(err, rows.Close()) }()

	itemsWithSaleCount := make([]*ItemWithSaleCount, 0)

	for rows.Next() {
		var id models.Id
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var itemCategoryId models.Id
		var sellerId models.Id
		var donation bool
		var charity bool
		var frozen bool
		var hidden bool
		var saleCount int

		err = rows.Scan(&id, &addedAt, &description, &priceInCents, &itemCategoryId, &sellerId, &donation, &charity, &frozen, &hidden, &saleCount)
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		itemWithSaleCount := ItemWithSaleCount{
			Item: models.Item{
				ItemID:       id,
				AddedAt:      addedAt,
				Description:  description,
				PriceInCents: priceInCents,
				CategoryID:   itemCategoryId,
				SellerID:     sellerId,
				Donation:     donation,
				Charity:      charity,
				Frozen:       frozen,
				Hidden:       hidden,
			},
			SaleCount: saleCount,
		}

		itemsWithSaleCount = append(itemsWithSaleCount, &itemWithSaleCount)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return itemsWithSaleCount, nil
}

// Returns the item with the given identifier.
// A NoSuchItemError is returned if no item with the given identifier exists.
func GetItemWithId(db *sql.DB, itemId models.Id) (*models.Item, error) {
	row := db.QueryRow(`
		SELECT added_at, description, price_in_cents, item_category_id, seller_id, donation, charity, frozen, hidden
		FROM items
		WHERE item_id = ?
	`, itemId)

	var addedAt models.Timestamp
	var description string
	var priceInCents models.MoneyInCents
	var categoryId models.Id
	var sellerId models.Id
	var donation bool
	var charity bool
	var frozen bool
	var hidden bool
	err := row.Scan(
		&addedAt,
		&description,
		&priceInCents,
		&categoryId,
		&sellerId,
		&donation,
		&charity,
		&frozen,
		&hidden,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dberr.ErrNoSuchItem
		}
		return nil, fmt.Errorf("failed to read row: %w", err)
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
	return &item, nil
}

// Returns all items with the given ids.
func GetItemsWithIds(db *sql.DB, itemIds []models.Id) (r_result map[models.Id]*models.Item, r_err error) {
	// Set up SQL query
	// Note that this does not detect nonexistent items, we deal with that later
	query := fmt.Sprintf(`
		SELECT item_id, added_at, description, price_in_cents, item_category_id, seller_id, donation, charity, frozen, hidden
		FROM items
		WHERE item_id IN (%s)
	`, placeholderString(len(itemIds)))
	convertedItemIds := algorithms.Map(itemIds, func(id models.Id) any { return id })
	rows, err := db.Query(query, convertedItemIds...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query to get items from database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	items := make(map[models.Id]*models.Item)
	for rows.Next() {
		var id models.Id
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var itemCategoryId models.Id
		var sellerId models.Id
		var donation bool
		var charity bool
		var frozen bool
		var hidden bool

		err = rows.Scan(&id, &addedAt, &description, &priceInCents, &itemCategoryId, &sellerId, &donation, &charity, &frozen, &hidden)
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		item := models.Item{
			ItemID:       id,
			AddedAt:      addedAt,
			Description:  description,
			PriceInCents: priceInCents,
			CategoryID:   itemCategoryId,
			SellerID:     sellerId,
			Donation:     donation,
			Charity:      charity,
			Frozen:       frozen,
			Hidden:       hidden,
		}
		items[id] = &item
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	// Check if all requested items were found
	if len(items) != len(itemIds) {
		for _, itemId := range itemIds {
			if _, ok := items[itemId]; !ok {
				return nil, fmt.Errorf("while getting items, among which %d: %w", itemId, dberr.ErrNoSuchItem)
			}
		}

		// If we get past the loop, it means that all items were found
		// but there were duplicates in the requested IDs, which is not an error
	}

	return items, nil
}

// Returns the total number of items in the database.
func CountItems(db *sql.DB, selection ItemSelection) (int, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(item_id)
		FROM %s
	`, ItemsTableFor(selection))
	row := db.QueryRow(query)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to read row: %w", err)
	}

	return count, nil
}

// AddItem adds an item to the database.
// An NoSuchUserError is returned if no user with the given sellerId exists.
// An ErrWrongRole is returned if sellerId does not refer to a seller.
// An NoSuchCategoryError is returned if the itemCategoryId is invalid.
// An InvalidPriceError is returned if the priceInCents is invalid.
func AddItem(
	db *sql.DB,
	addedAt models.Timestamp,
	description string,
	priceInCents models.MoneyInCents,
	itemCategoryId models.Id,
	sellerId models.Id,
	donation bool,
	charity bool,
	frozen bool,
	hidden bool) (models.Id, error) {

	if !models.IsValidPrice(priceInCents) {
		return 0, fmt.Errorf("failed to add item with price %d: %w", priceInCents, dberr.ErrInvalidPrice)
	}
	if !models.IsValidItemDescription(description) {
		return 0, fmt.Errorf("failed to add item with description %s: %w", description, dberr.ErrInvalidItemDescription)
	}
	if err := EnsureUserExistsAndHasRole(db, sellerId, models.NewSellerRoleId()); err != nil {
		return 0, fmt.Errorf("could not ensure user %d exists and is seller: %w", sellerId, err)
	}
	if frozen && hidden {
		return 0, fmt.Errorf("failed to add item: %w", dberr.ErrHiddenFrozenItem)
	}

	result, err := db.Exec(
		`
			INSERT INTO items (added_at, description, price_in_cents, item_category_id, seller_id, donation, charity, frozen, hidden)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`,
		addedAt,
		description,
		priceInCents,
		itemCategoryId,
		sellerId,
		donation,
		charity,
		frozen,
		hidden,
	)
	if err != nil {
		categoryExists, err2 := CategoryWithIdExists(db, itemCategoryId)
		if err2 != nil {
			return 0, fmt.Errorf("failed ot determine whether category with given id exists: %w", err)
		}

		if !categoryExists {
			return 0, fmt.Errorf("failed to add item with category %d: %w", itemCategoryId, dberr.ErrNoSuchCategory)
		}

		return 0, fmt.Errorf("failed to insert item: %w", err)
	}

	itemId, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to determine id of inserted item: %w", err)
	}

	return models.Id(itemId), nil
}

// Returns true if an item with the given identifier exists in the database.
func ItemWithIdExists(db *sql.DB, itemId models.Id) (bool, error) {
	row := db.QueryRow(
		`
			SELECT 1
			FROM items
			WHERE item_id = $1
		`,
		itemId,
	)

	var result int
	err := row.Scan(&result)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func ItemsExist(db QueryHandler, itemIds []models.Id) (bool, error) {
	if len(itemIds) == 0 {
		return true, nil
	}

	itemIds = algorithms.RemoveDuplicates(itemIds)

	// Set up SQL query
	query := fmt.Sprintf(`
		SELECT COUNT(item_id)
		FROM items
		WHERE item_id IN (%s)
	`, placeholderString(len(itemIds)))

	convertedItemIds := algorithms.Map(itemIds, func(id models.Id) any { return id })
	row := db.QueryRow(query, convertedItemIds...)

	var count int
	err := row.Scan(&count)

	if err != nil {
		return false, err
	}

	return count == len(itemIds), nil
}

// EnsureItemsExist checks if all items with the given IDs exist in the database.
// If any item does not exist, it returns a NoSuchItemError.
func EnsureItemsExist(db QueryHandler, itemIds []models.Id) error {
	itemsExist, err := ItemsExist(db, itemIds)
	if err != nil {
		return fmt.Errorf("failed to ensure items exist: %w", err)
	}
	if !itemsExist {
		return fmt.Errorf("failed to ensure items exist: %w", dberr.ErrNoSuchItem)
	}

	return nil
}

func UpdateFreezeStatusOfItems(db *sql.DB, itemIds []models.Id, frozen bool) (r_err error) {
	if len(itemIds) == 0 {
		return nil
	}

	itemIds = algorithms.RemoveDuplicates(itemIds)
	convertedItemIds := algorithms.Map(itemIds, func(id models.Id) any { return id })

	transaction, err := NewTransaction(db)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, transaction.Rollback()) }()

	if err := EnsureItemsExist(transaction, itemIds); err != nil {
		return err
	}

	if err := EnsureNoHiddenItems(transaction, itemIds); err != nil {
		return fmt.Errorf("failed to ensure no hidden items: %w", err)
	}

	query := fmt.Sprintf(`
		UPDATE items
		SET frozen = ?
		WHERE item_id IN (%s)
	`, placeholderString(len(itemIds)))

	sqlValues := append([]any{frozen}, convertedItemIds...)

	if _, err := transaction.Exec(query, sqlValues...); err != nil {
		return err
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func UpdateHiddenStatusOfItems(db *sql.DB, itemIds []models.Id, hidden bool) (r_err error) {
	if len(itemIds) == 0 {
		return nil
	}

	itemIds = algorithms.RemoveDuplicates(itemIds)
	convertedItemIds := algorithms.Map(itemIds, func(id models.Id) any { return id })

	transaction, err := NewTransaction(db)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, transaction.Rollback()) }()

	if err := EnsureItemsExist(transaction, itemIds); err != nil {
		return err
	}

	// Check if none of the items are frozen
	if err := EnsureNoFrozenItems(transaction, itemIds); err != nil {
		return err
	}

	query := fmt.Sprintf(`
		UPDATE items
		SET hidden = ?
		WHERE item_id IN (%s)
	`, placeholderString(len(itemIds)))

	sqlValues := append([]any{hidden}, convertedItemIds...)

	if _, err := transaction.Exec(query, sqlValues...); err != nil {
		return err
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func partitionItemsBy(db QueryHandler, itemIds []models.Id, columnName string) (*algorithms.Set[models.Id], *algorithms.Set[models.Id], error) {
	query := fmt.Sprintf(`
		SELECT item_id, %s
		FROM items
		WHERE item_id IN (%s)
	`, columnName, placeholderString(len(itemIds)))
	convertedItemIds := algorithms.Map(itemIds, func(id models.Id) any { return id })
	rows, err := db.Query(query, convertedItemIds...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()

	falseSet := algorithms.NewSet[models.Id]()
	trueSet := algorithms.NewSet[models.Id]()
	for rows.Next() {
		var id models.Id
		var hiddenStatus bool

		err = rows.Scan(&id, &hiddenStatus)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan items: %w", err)
		}

		if hiddenStatus {
			trueSet.Add(id)
		} else {
			falseSet.Add(id)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return &falseSet, &trueSet, nil
}

// PartitionItemsByHiddenStatus partitions the given item IDs into two sets: one for unhidden items and one for hidden items.
// If an item ID does not exist in the database, it is ignored.
func PartitionItemsByHiddenStatus(db QueryHandler, itemIds []models.Id) (*algorithms.Set[models.Id], *algorithms.Set[models.Id], error) {
	return partitionItemsBy(db, itemIds, "hidden")
}

// PartitionItemsByFrozenStatus partitions the given item IDs into two sets: one for nonfrozen items and one for frozen items.
// If an item ID does not exist in the database, it is ignored.
func PartitionItemsByFrozenStatus(db QueryHandler, itemIds []models.Id) (*algorithms.Set[models.Id], *algorithms.Set[models.Id], error) {
	return partitionItemsBy(db, itemIds, "frozen")
}

// ContainsHiddenItems checks if any of the given items are hidden.
// It returns true if at least one item is hidden, and false otherwise.
// It is not an error when nonexistent items are passed in, they are simply ignored.
func ContainsHiddenItems(qh QueryHandler, itemIds []models.Id) (bool, error) {
	if len(itemIds) == 0 {
		return false, nil
	}

	_, hidden, err := PartitionItemsByHiddenStatus(qh, itemIds)
	if err != nil {
		return false, err
	}

	containsHidden := hidden.Len() > 0
	return containsHidden, nil
}

// ContainsFrozenItems checks if any of the given items are frozen.
// It returns true if at least one item is frozen, and false otherwise.
// It is not an error when nonexistent items are passed in, they are simply ignored.
func ContainsFrozenItems(qh QueryHandler, itemIds []models.Id) (bool, error) {
	if len(itemIds) == 0 {
		return false, nil
	}

	_, frozen, err := PartitionItemsByFrozenStatus(qh, itemIds)
	if err != nil {
		return false, err
	}

	containsFrozen := frozen.Len() > 0
	return containsFrozen, nil
}

// IsItemFrozen checks if none of the items is frozen.
func EnsureNoFrozenItems(qh QueryHandler, itemIds []models.Id) error {
	containsFrozen, err := ContainsFrozenItems(qh, itemIds)

	if err != nil {
		return fmt.Errorf("failed to check for frozen items: %w", err)
	}

	if containsFrozen {
		return dberr.ErrItemFrozen
	}

	return nil
}

// EnsureNoHiddenItems checks if none of the items is hidden.
func EnsureNoHiddenItems(qh QueryHandler, itemIds []models.Id) error {
	containsHidden, err := ContainsHiddenItems(qh, itemIds)

	if err != nil {
		return fmt.Errorf("failed to check for hidden items: %w", err)
	}

	if containsHidden {
		return dberr.ErrItemHidden
	}

	return nil
}

func IsItemFrozen(db *sql.DB, itemId models.Id) (bool, error) {
	nonfrozen, frozen, err := PartitionItemsByFrozenStatus(db, []models.Id{itemId})
	if err != nil {
		return false, err
	}

	isFrozen := frozen.Len() > 0
	if isFrozen {
		return true, nil
	}

	isUnfrozen := nonfrozen.Len() > 0
	if isUnfrozen {
		return false, nil
	}

	return false, fmt.Errorf("failed to check if item %d is frozen: %w", itemId, dberr.ErrNoSuchItem)
}

func IsItemHidden(db *sql.DB, itemId models.Id) (bool, error) {
	unhidden, hidden, err := PartitionItemsByHiddenStatus(db, []models.Id{itemId})
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

	return false, fmt.Errorf("failed to check if item %d is hidden: %w", itemId, dberr.ErrNoSuchItem)
}

func RemoveItemWithId(db *sql.DB, itemId models.Id) error {
	itemExists, err := ItemWithIdExists(db, itemId)

	if err != nil {
		return err
	}

	if !itemExists {
		return fmt.Errorf("failed to remove item with id %d: %w", itemId, dberr.ErrNoSuchItem)
	}

	_, err = db.Exec(
		`
			DELETE FROM items
			WHERE item_id = $1
		`,
		itemId,
	)

	return err
}

type ItemUpdate struct {
	AddedAt      *models.Timestamp
	Description  *string
	PriceInCents *models.MoneyInCents
	CategoryId   *models.Id
	Donation     *bool
	Charity      *bool
}

func UpdateItem(db *sql.DB, itemId models.Id, itemUpdate *ItemUpdate) error {
	if itemUpdate == nil {
		slog.Error("parameter itemUpdate is nil")
		os.Exit(1)
	}

	item, err := GetItemWithId(db, itemId)
	if err != nil {
		return err
	}

	if item.Frozen {
		return dberr.ErrItemFrozen
	}

	if item.Hidden {
		return dberr.ErrItemHidden
	}

	sqlUpdates := []string{}
	sqlValues := []any{}

	if itemUpdate.AddedAt != nil {
		sqlUpdates = append(sqlUpdates, "added_at = ?")
		sqlValues = append(sqlValues, *itemUpdate.AddedAt)
	}

	if itemUpdate.Description != nil {
		sqlUpdates = append(sqlUpdates, "description = ?")
		sqlValues = append(sqlValues, *itemUpdate.Description)
	}

	if itemUpdate.PriceInCents != nil {
		if !models.IsValidPrice(*itemUpdate.PriceInCents) {
			return fmt.Errorf("failed to updated item's price to %d: %w", *itemUpdate.PriceInCents, dberr.ErrInvalidPrice)
		}

		sqlUpdates = append(sqlUpdates, "price_in_cents = ?")
		sqlValues = append(sqlValues, *itemUpdate.PriceInCents)
	}

	if itemUpdate.CategoryId != nil {
		categoryExists, err := CategoryWithIdExists(db, *itemUpdate.CategoryId)
		if err != nil {
			return err
		}

		if !categoryExists {
			return fmt.Errorf("failed to update item's category to %d", *itemUpdate.CategoryId)
		}

		sqlUpdates = append(sqlUpdates, "item_category_id = ?")
		sqlValues = append(sqlValues, *itemUpdate.CategoryId)
	}

	if itemUpdate.Donation != nil {
		sqlUpdates = append(sqlUpdates, "donation = ?")
		sqlValues = append(sqlValues, *itemUpdate.Donation)
	}

	if itemUpdate.Charity != nil {
		sqlUpdates = append(sqlUpdates, "charity = ?")
		sqlValues = append(sqlValues, *itemUpdate.Charity)
	}

	if len(sqlUpdates) == 0 {
		return nil
	}

	sqlValues = append(sqlValues, itemId)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE item_id = ?", "items", strings.Join(sqlUpdates, ", "))

	if _, err := db.Exec(query, sqlValues...); err != nil {
		return err
	}

	return nil
}
