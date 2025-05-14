package queries

import (
	"bctbackend/algorithms"
	"bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

func GetItems(db *sql.DB, receiver func(*models.Item) error) error {
	rows, err := db.Query(`
		SELECT item_id, added_at, description, price_in_cents, item_category_id, seller_id, donation, charity, frozen
		FROM items
		ORDER BY item_id ASC
	`)
	if err != nil {
		return err
	}

	defer func() { err = errors.Join(err, rows.Close()) }()

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

		err = rows.Scan(&id, &addedAt, &description, &priceInCents, &itemCategoryId, &sellerId, &donation, &charity, &frozen)

		if err != nil {
			return err
		}

		item := models.NewItem(id, addedAt, description, priceInCents, itemCategoryId, sellerId, donation, charity, frozen)

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
func GetSellerItems(db *sql.DB, sellerId models.Id) (r_items []*models.Item, r_err error) {
	if err := CheckUserRole(db, sellerId, models.SellerRoleId); err != nil {
		return nil, err
	}

	rows, err := db.Query(
		`
			SELECT item_id, added_at, description, price_in_cents, item_category_id, seller_id, donation, charity, frozen
			FROM items
			WHERE seller_id = ?
			ORDER BY added_at, item_id ASC
		`,
		sellerId,
	)
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

		err = rows.Scan(&id, &addedAt, &description, &priceInCents, &itemCategoryId, &sellerId, &donation, &charity, &frozen)
		if err != nil {
			return nil, err
		}

		item := models.NewItem(id, addedAt, description, priceInCents, itemCategoryId, sellerId, donation, charity, frozen)

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
			SELECT items.item_id, added_at, description, price_in_cents, item_category_id, seller_id, donation, charity, frozen, COALESCE(COUNT(sale_items.sale_id), 0) AS sale_count
			FROM items LEFT JOIN sale_items ON items.item_id = sale_items.item_id
			WHERE seller_id = ?
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
		var saleCount int

		err = rows.Scan(&id, &addedAt, &description, &priceInCents, &itemCategoryId, &sellerId, &donation, &charity, &frozen, &saleCount)
		if err != nil {
			return nil, err
		}

		item := ItemWithSaleCount{
			Item:      *models.NewItem(id, addedAt, description, priceInCents, itemCategoryId, sellerId, donation, charity, frozen),
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
		SELECT added_at, description, price_in_cents, item_category_id, seller_id, donation, charity, frozen
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
	err := row.Scan(&addedAt, &description, &priceInCents, &categoryId, &sellerId, &donation, &charity, &frozen)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, &NoSuchItemError{Id: &itemId}
	}

	if err != nil {
		return nil, err
	}

	item := models.NewItem(itemId, addedAt, description, priceInCents, categoryId, sellerId, donation, charity, frozen)

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
		SELECT item_id, added_at, description, price_in_cents, item_category_id, seller_id, donation, charity, frozen
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

		err = rows.Scan(&id, &addedAt, &description, &priceInCents, &itemCategoryId, &sellerId, &donation, &charity, &frozen)
		if err != nil {
			return nil, err
		}

		item := models.NewItem(id, addedAt, description, priceInCents, itemCategoryId, sellerId, donation, charity, frozen)
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
func CountItems(db *sql.DB) (int, error) {
	row := db.QueryRow(`
		SELECT COUNT(item_id)
		FROM items
	`)

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

func CheckItemsExistence(db *sql.DB, itemIds []models.Id) (bool, error) {
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

	transaction, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	transactionCommitted := false
	defer func() {
		if !transactionCommitted {
			r_err = errors.Join(r_err, transaction.Rollback())
		}
	}()

	// Check if all items exist and none are hidden
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

	// Signals that no rollback is needed
	transactionCommitted = true

	return nil
}

func UpdateHiddenStatusOfItems(db *sql.DB, itemIds []models.Id, hidden bool) (r_err error) {
	if len(itemIds) == 0 {
		return nil
	}

	itemIds = algorithms.RemoveDuplicates(itemIds)
	convertedItemIds := algorithms.Map(itemIds, func(id models.Id) any { return id })

	transaction, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	transactionCommitted := false
	defer func() {
		if !transactionCommitted {
			r_err = errors.Join(r_err, transaction.Rollback())
		}
	}()

	// Check if all items exist and none are frozen
	containsFrozen, err := ContainsFrozenItems(transaction, itemIds)
	if err != nil {
		return fmt.Errorf("failed to check for frozen items: %w", err)
	}
	if containsFrozen {
		return &ItemHiddenError{}
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

	// Signals that no rollback is needed
	transactionCommitted = true

	return nil
}

func ContainsHiddenItems(qh QueryHandler, itemIds []models.Id) (r_result bool, r_err error) {
	if len(itemIds) == 0 {
		return false, nil
	}

	itemIds = algorithms.RemoveDuplicates(itemIds)

	query := fmt.Sprintf(`
		SELECT hidden, COUNT(item_id)
		FROM items
		WHERE item_id IN (%s)
		GROUP BY hidden
	`, placeholderString(len(itemIds)))

	convertedItemIds := algorithms.Map(itemIds, func(id models.Id) any { return id })
	rows, err := qh.Query(query, convertedItemIds...)
	if err != nil {
		return false, fmt.Errorf("failed to query items: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	totalCount := 0
	hiddenFound := false
	for rows.Next() {
		var hidden bool
		var count int

		err = rows.Scan(&hidden, &count)
		if err != nil {
			return false, fmt.Errorf("failed to scan items: %w", err)
		}

		if hidden && count > 0 {
			hiddenFound = true
		}

		totalCount += count
	}

	if totalCount != len(itemIds) {
		return false, &NoSuchItemError{Id: nil}
	}

	return hiddenFound, nil
}

func ContainsFrozenItems(qh QueryHandler, itemIds []models.Id) (r_result bool, r_err error) {
	if len(itemIds) == 0 {
		return false, nil
	}

	itemIds = algorithms.RemoveDuplicates(itemIds)

	query := fmt.Sprintf(`
		SELECT frozen, COUNT(item_id)
		FROM items
		WHERE item_id IN (%s)
		GROUP BY frozen
	`, placeholderString(len(itemIds)))

	convertedItemIds := algorithms.Map(itemIds, func(id models.Id) any { return id })
	rows, err := qh.Query(query, convertedItemIds...)
	if err != nil {
		return false, fmt.Errorf("failed to query items: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	totalCount := 0
	frozenFound := false
	for rows.Next() {
		var hidden bool
		var count int

		err = rows.Scan(&hidden, &count)
		if err != nil {
			return false, fmt.Errorf("failed to scan items: %w", err)
		}

		if hidden && count > 0 {
			frozenFound = true
		}

		totalCount += count
	}

	if totalCount != len(itemIds) {
		return false, &NoSuchItemError{Id: nil}
	}

	return frozenFound, nil
}

func IsItemFrozen(db *sql.DB, itemId models.Id) (bool, error) {
	row := db.QueryRow(
		`
			SELECT frozen
			FROM items
			WHERE item_id = $1
		`,
		itemId,
	)

	var isFrozen bool
	err := row.Scan(&isFrozen)

	if errors.Is(err, sql.ErrNoRows) {
		return false, &NoSuchItemError{Id: &itemId}
	}

	if err != nil {
		return false, err
	}

	return isFrozen, nil
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
