package queries

import (
	"bctbackend/database/models"
	"database/sql"
	"errors"
)

func GetItems(db *sql.DB, receiver func(*models.Item) error) error {
	rows, err := db.Query(`
		SELECT item_id, added_at, description, price_in_cents, item_category_id, seller_id, donation, charity
		FROM items
		ORDER BY item_id ASC
	`)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var id models.Id
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var itemCategoryId models.Id
		var sellerId models.Id
		var donation bool
		var charity bool

		err = rows.Scan(&id, &addedAt, &description, &priceInCents, &itemCategoryId, &sellerId, &donation, &charity)

		if err != nil {
			return err
		}

		item := models.NewItem(id, addedAt, description, priceInCents, itemCategoryId, sellerId, donation, charity)

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
func GetSellerItems(db *sql.DB, sellerId models.Id) ([]*models.Item, error) {
	if err := CheckUserRole(db, sellerId, models.SellerRoleId); err != nil {
		return nil, err
	}

	rows, err := db.Query(
		`
			SELECT item_id, added_at, description, price_in_cents, item_category_id, seller_id, donation, charity
			FROM items
			WHERE seller_id = ?
			ORDER BY added_at, item_id ASC
		`,
		sellerId,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

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

		err = rows.Scan(&id, &addedAt, &description, &priceInCents, &itemCategoryId, &sellerId, &donation, &charity)

		if err != nil {
			return nil, err
		}

		item := models.NewItem(id, addedAt, description, priceInCents, itemCategoryId, sellerId, donation, charity)

		items = append(items, item)
	}

	return items, nil
}

// Returns the item with the given identifier.
// A NoSuchItemError is returned if no item with the given identifier exists.
func GetItemWithId(db *sql.DB, itemId models.Id) (*models.Item, error) {
	row := db.QueryRow(`
		SELECT item_id, added_at, description, price_in_cents, item_category_id, seller_id, donation, charity
		FROM items
		WHERE item_id = ?
	`, itemId)

	var id models.Id
	var addedAt models.Timestamp
	var description string
	var priceInCents models.MoneyInCents
	var itemCategoryId models.Id
	var sellerId models.Id
	var donation bool
	var charity bool

	err := row.Scan(&id, &addedAt, &description, &priceInCents, &itemCategoryId, &sellerId, &donation, &charity)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, &NoSuchItemError{Id: itemId}
	}

	if err != nil {
		return nil, err
	}

	item := models.NewItem(id, addedAt, description, priceInCents, itemCategoryId, sellerId, donation, charity)

	return item, nil
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
	charity bool) (models.Id, error) {

	if priceInCents <= 0 {
		return 0, &InvalidPriceError{PriceInCents: priceInCents}
	}

	if err := CheckUserRole(db, sellerId, models.SellerRoleId); err != nil {
		return 0, err
	}

	result, err := db.Exec(
		`
			INSERT INTO items (added_at, description, price_in_cents, item_category_id, seller_id, donation, charity)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		addedAt,
		description,
		priceInCents,
		itemCategoryId,
		sellerId,
		donation,
		charity)

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
