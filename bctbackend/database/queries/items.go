package queries

import (
	"bctbackend/algorithms"
	"bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func GetItems(db *sql.DB, receiver func(*models.Item) error, includeHidden bool) error {
	// Build SQL query
	var whereClause string
	if includeHidden {
		whereClause = ""
	} else {
		whereClause = "WHERE hidden = false"
	}
	query := fmt.Sprintf(`
		SELECT item_id, added_at, description, price_in_cents, item_category_id, seller_id, donation, charity, frozen, hidden
		FROM items
		%s
		ORDER BY item_id ASC
	`, whereClause)

	// Perform query
	rows, err := db.Query(query)
	if err != nil {
		return err
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
			return err
		}

		item := models.NewItem(id, addedAt, description, priceInCents, itemCategoryId, sellerId, donation, charity, frozen, hidden)

		if err := receiver(item); err != nil {
			return err
		}
	}

	return nil
}

// Returns the items associated with the given seller.
// The items are ordered by their time of addition, then by id.
// An NoSuchUserError is returned if no user with the given sellerId exists.
// An InvalidRoleError is returned if sellerId does not refer to a seller.
func GetSellerItems(db *sql.DB, sellerId models.Id, includeHidden bool) (r_items []*models.Item, r_err error) {
	// Ensure that sellerId is associated with a seller
	if err := CheckUserRole(db, sellerId, models.SellerRoleId); err != nil {
		return nil, err
	}

	// Build SQL query
	whereClause := "WHERE seller_id = ?"

	if !includeHidden {
		whereClause += " AND hidden = false"
	}

	query := fmt.Sprintf(`
		SELECT item_id, added_at, description, price_in_cents, item_category_id, seller_id, donation, charity, frozen, hidden
		FROM items
		%s
		ORDER BY added_at, item_id ASC
	`, whereClause)

	rows, err := db.Query(query, sellerId)
	if err != nil {
		return nil, err
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
			return nil, err
		}

		item := models.NewItem(id, addedAt, description, priceInCents, itemCategoryId, sellerId, donation, charity, frozen, hidden)

		items = append(items, item)
	}

	return items, nil
}

type ItemWithSaleCount struct {
	models.Item
	SaleCount int
}

// Returns the items associated with the given seller.
// The items are ordered by their time of addition, then by id.
// An NoSuchUserError is returned if no user with the given sellerId exists.
// An InvalidRoleError is returned if sellerId does not refer to a seller.
func GetSellerItemsWithSaleCounts(db *sql.DB, sellerId models.Id) (r_items []*ItemWithSaleCount, r_err error) {
	if err := CheckUserRole(db, sellerId, models.SellerRoleId); err != nil {
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
		return nil, fmt.Errorf("error occurred while getting seller items with sale counts: %w", err)
	}

	defer func() { err = errors.Join(err, rows.Close()) }()

	items := make([]*ItemWithSaleCount, 0)

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
			return nil, err
		}

		item := ItemWithSaleCount{
			Item:      *models.NewItem(id, addedAt, description, priceInCents, itemCategoryId, sellerId, donation, charity, frozen, hidden),
			SaleCount: saleCount,
		}

		items = append(items, &item)
	}

	err = nil
	return items, err
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
	err := row.Scan(&addedAt, &description, &priceInCents, &categoryId, &sellerId, &donation, &charity, &frozen, &hidden)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, &NoSuchItemError{Id: &itemId}
	}

	if err != nil {
		return nil, err
	}

	item := models.NewItem(itemId, addedAt, description, priceInCents, categoryId, sellerId, donation, charity, frozen, hidden)

	return item, nil
}

// Returns all items with the given ids.
func GetItemsWithIds(db *sql.DB, itemIds []models.Id) (map[models.Id]*models.Item, error) {
	// Handle the special case of zero items efficiently
	if len(itemIds) == 0 {
		return nil, nil
	}

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
		return nil, err
	}
	defer rows.Close()

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
			return nil, err
		}

		item := models.NewItem(id, addedAt, description, priceInCents, itemCategoryId, sellerId, donation, charity, frozen, hidden)
		items[id] = item
	}

	// Check if all requested items were found
	if len(items) != len(itemIds) {
		for _, itemId := range itemIds {
			if _, ok := items[itemId]; !ok {
				return nil, &NoSuchItemError{Id: &itemId}
			}
		}

		// If we get past the loop, it means that all items were found
		// but there were duplicates in the requested IDs, which is not an error
	}

	return items, nil
}

// Returns the total number of items in the database.
func CountItems(db *sql.DB, includeHidden bool) (int, error) {
	query := `
		SELECT COUNT(item_id)
		FROM items
	`

	if !includeHidden {
		query += " WHERE hidden = false"
	}

	row := db.QueryRow(query)

	var count int
	err := row.Scan(&count)

	return count, err
}

// AddItem adds an item to the database.
// An NoSuchUserError is returned if no user with the given sellerId exists.
// An InvalidRoleError is returned if sellerId does not refer to a seller.
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
		return 0, &InvalidPriceError{PriceInCents: priceInCents}
	}

	if !models.IsValidItemDescription(description) {
		return 0, &InvalidItemDescriptionError{Description: description}
	}

	if err := CheckUserRole(db, sellerId, models.SellerRoleId); err != nil {
		return 0, err
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
			return 0, err
		}

		if !categoryExists {
			return 0, &NoSuchCategoryError{CategoryId: itemCategoryId}
		}

		return 0, err
	}

	itemId, err := result.LastInsertId()

	if err != nil {
		return 0, err
	}

	return itemId, nil
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
	defer func() { transaction.Rollback() }()

	itemsExist, err := ItemsExist(transaction, itemIds)
	if err != nil {
		return err
	}
	if !itemsExist {
		return &NoSuchItemError{Id: nil}
	}

	containsHidden, err := ContainsHiddenItems(transaction, itemIds)
	if err != nil {
		return fmt.Errorf("failed to check for hidden items: %w", err)
	}
	if containsHidden {
		return &ItemHiddenError{}
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
	defer func() { transaction.Rollback() }()

	// Check if all items exist
	itemsExist, err := ItemsExist(transaction, itemIds)
	if err != nil {
		return err
	}
	if !itemsExist {
		return &NoSuchItemError{Id: nil}
	}

	// Check if none of the items are frozen
	containsFrozen, err := ContainsFrozenItems(transaction, itemIds)
	if err != nil {
		return fmt.Errorf("failed to check for frozen items: %w", err)
	}
	if containsFrozen {
		return &ItemFrozenError{}
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

// PartitionItemsByHiddenStatus partitions the given item IDs into two sets: one for unhidden items and one for hidden items.
// If an item ID does not exist in the database, it is ignored.
func PartitionItemsByHiddenStatus(db QueryHandler, itemIds []models.Id) (*algorithms.Set[models.Id], *algorithms.Set[models.Id], error) {
	query := fmt.Sprintf(`
		SELECT item_id, hidden
		FROM items
		WHERE item_id IN (%s)
	`, placeholderString(len(itemIds)))
	convertedItemIds := algorithms.Map(itemIds, func(id models.Id) any { return id })
	rows, err := db.Query(query, convertedItemIds...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()

	unhidden := algorithms.NewSet[models.Id]()
	hidden := algorithms.NewSet[models.Id]()
	for rows.Next() {
		var id models.Id
		var hiddenStatus bool

		err = rows.Scan(&id, &hiddenStatus)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan items: %w", err)
		}

		if hiddenStatus {
			hidden.Add(id)
		} else {
			unhidden.Add(id)
		}
	}

	return unhidden, hidden, nil
}

// PartitionItemsByFrozenStatus partitions the given item IDs into two sets: one for nonfrozen items and one for frozen items.
// If an item ID does not exist in the database, it is ignored.
func PartitionItemsByFrozenStatus(db QueryHandler, itemIds []models.Id) (*algorithms.Set[models.Id], *algorithms.Set[models.Id], error) {
	query := fmt.Sprintf(`
		SELECT item_id, frozen
		FROM items
		WHERE item_id IN (%s)
	`, placeholderString(len(itemIds)))
	convertedItemIds := algorithms.Map(itemIds, func(id models.Id) any { return id })
	rows, err := db.Query(query, convertedItemIds...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()

	nonfrozen := algorithms.NewSet[models.Id]()
	frozen := algorithms.NewSet[models.Id]()
	for rows.Next() {
		var id models.Id
		var frozenStatus bool

		err = rows.Scan(&id, &frozenStatus)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan items: %w", err)
		}

		if frozenStatus {
			frozen.Add(id)
		} else {
			nonfrozen.Add(id)
		}
	}

	return nonfrozen, frozen, nil
}

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

func ContainsFrozenItems(qh QueryHandler, itemIds []models.Id) (r_result bool, r_err error) {
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

	return false, &NoSuchItemError{Id: &itemId}
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

	return false, &NoSuchItemError{Id: &itemId}
}

func RemoveItemWithId(db *sql.DB, itemId models.Id) error {
	itemExists, err := ItemWithIdExists(db, itemId)

	if err != nil {
		return err
	}

	if !itemExists {
		return &NoSuchItemError{Id: &itemId}
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
		return fmt.Errorf("bug: itemUpdate is nil")
	}

	item, err := GetItemWithId(db, itemId)
	if err != nil {
		return err
	}

	if item.Frozen {
		return &ItemFrozenError{Id: itemId}
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
			return &InvalidPriceError{PriceInCents: *itemUpdate.PriceInCents}
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
			return &NoSuchCategoryError{CategoryId: *itemUpdate.CategoryId}
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
