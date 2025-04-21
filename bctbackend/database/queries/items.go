package queries

import (
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
	SaleCount int64
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
			SELECT item_id, added_at, description, price_in_cents, item_category_id, seller_id, donation, charity, frozen, COALESCE(COUNT(sales.sale_id), 0) AS sale_count
			FROM items LEFT JOIN sales ON items.item_id = sales.item_id
			WHERE seller_id = ?
			ORDER BY added_at, item_id ASC
		`,
		sellerId,
	)
	if err != nil {
		return nil, err
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
		var saleCount int64

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

	item := models.Item{ItemId: itemId}
	err := row.Scan(&item.AddedAt, &item.Description, &item.PriceInCents, &item.CategoryId, &item.SellerId, &item.Donation, &item.Charity, &item.Frozen)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, &NoSuchItemError{Id: itemId}
	}

	if err != nil {
		return nil, err
	}

	return &item, nil
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
	frozen bool) (models.Id, error) {

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
			INSERT INTO items (added_at, description, price_in_cents, item_category_id, seller_id, donation, charity, frozen)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`,
		addedAt,
		description,
		priceInCents,
		itemCategoryId,
		sellerId,
		donation,
		charity,
		frozen)

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

func FreezeItem(db *sql.DB, itemId models.Id) error {
	itemExists, err := ItemWithIdExists(db, itemId)
	if err != nil {
		return err
	}
	if !itemExists {
		return &NoSuchItemError{Id: itemId}
	}

	_, err = db.Exec(
		`
			UPDATE items
			SET frozen = TRUE
			WHERE item_id = $1
		`,
		itemId,
	)
	if err != nil {
		return err
	}

	return nil
}

func ItemWithIdIsFrozen(db *sql.DB, itemId models.Id) (bool, error) {
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
		return false, &NoSuchItemError{Id: itemId}
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
		return &NoSuchItemError{Id: itemId}
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
