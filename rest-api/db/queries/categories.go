package queries

import (
	models "bctrest/db/models"
	"database/sql"
)

func CategoryWithIdExists(
	db *sql.DB,
	categoryId models.Id) bool {

	row := db.QueryRow(
		`
			SELECT 1
			FROM item_categories
			WHERE item_category_id = $1
		`,
		categoryId,
	)

	err := row.Scan()

	return err == nil
}
