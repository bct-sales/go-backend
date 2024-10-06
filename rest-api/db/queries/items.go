package queries

import (
	models "bctrest/db/models"
	"database/sql"
)

func GetItems(db *sql.DB) ([]models.Item, error) {
	rows, err := db.Query(`
		SELECT item_id, timestamp, description, price_in_cents, item_category_id, owner_id, recipient_id, charity
		FROM items
	`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	items := make([]models.Item, 0)

	for rows.Next() {
		var id models.Id
		var timestamp models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var itemCategoryId models.Id
		var ownerId models.Id
		var recipientId models.Id
		var charity bool

		err = rows.Scan(&id, &timestamp, &description, &priceInCents, &itemCategoryId, &ownerId, &recipientId, &charity)

		if err != nil {
			return nil, err
		}

		item := models.NewItem(id, timestamp, description, priceInCents, itemCategoryId, ownerId, recipientId, charity)

		items = append(items, *item)
	}

	return items, nil
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
	timestamp models.Timestamp,
	description string,
	priceInCents models.MoneyInCents,
	itemCategoryId models.Id,
	ownerId models.Id,
	recipientId models.Id,
	charity bool) error {

	_, err := db.Exec(
		`
			INSERT INTO items (timestamp, description, price_in_cents, item_category_id, owner_id, recipient_id, charity)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		timestamp,
		description,
		priceInCents,
		itemCategoryId,
		ownerId,
		recipientId,
		charity)

	return err
}

func ItemWithIdExists(
	db *sql.DB,
	itemId models.Id) bool {

	row := db.QueryRow(
		`
			SELECT 1
			FROM items
			WHERE item_id = $1
		`,
		itemId,
	)

	err := row.Scan()

	return err == nil
}
