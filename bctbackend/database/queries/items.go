package queries

import (
	models "bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"
)

func GetItems(db *sql.DB) ([]models.Item, error) {
	rows, err := db.Query(`
		SELECT item_id, added_at, description, price_in_cents, item_category_id, seller_id, donation, charity
		FROM items
		ORDER BY item_id ASC
	`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := make([]models.Item, 0)

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

		items = append(items, *item)
	}

	return items, nil
}

func GetSellerItems(db *sql.DB, sellerId models.Id) ([]models.Item, error) {
	rows, err := db.Query(
		`
			SELECT item_id, added_at, description, price_in_cents, item_category_id, seller_id, donation, charity
			FROM items
			WHERE seller_id = ?
			ORDER BY item_id ASC
		`,
		sellerId,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := make([]models.Item, 0)

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

		items = append(items, *item)
	}

	return items, nil
}

type ItemNotFoundError struct {
	Id models.Id
}

func (e *ItemNotFoundError) Error() string {
	return fmt.Sprintf("item with id %d not found", e.Id)
}

func (e *ItemNotFoundError) Unwrap() error {
	return nil
}

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
		return nil, &ItemNotFoundError{Id: itemId}
	}

	if err != nil {
		return nil, err
	}

	item := models.NewItem(id, addedAt, description, priceInCents, itemCategoryId, sellerId, donation, charity)

	return item, nil
}

func CountItems(db *sql.DB) (int, error) {
	row := db.QueryRow(`
		SELECT COUNT(item_id)
		FROM items
	`)

	var count int
	err := row.Scan(&count)

	return count, err
}

func AddItem(
	db *sql.DB,
	addedAt models.Timestamp,
	description string,
	priceInCents models.MoneyInCents,
	itemCategoryId models.Id,
	sellerId models.Id,
	donation bool,
	charity bool) (models.Id, error) {

	err := CheckUserRole(db, sellerId, models.SellerRoleId)

	if err != nil {
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
		return 0, err
	}

	return result.LastInsertId()
}

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
		return &ItemNotFoundError{Id: itemId}
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
