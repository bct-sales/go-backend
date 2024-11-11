package queries

import (
	models "bctbackend/database/models"
	"database/sql"
	"errors"
)

type SaleItemInformation struct {
	Description    string
	ItemCategoryId models.Id
	PriceInCents   models.MoneyInCents
	SellCount      int64
}

func GetSaleItemInformation(
	db *sql.DB,
	itemId models.Id) (*SaleItemInformation, error) {

	row := db.QueryRow(
		`
			SELECT description, price_in_cents, item_category_id, COUNT(si.sale_id)
			FROM items i LEFT JOIN sale_items si ON i.item_id = si.item_id
			GROUP BY i.item_id
			HAVING i.item_id = ?
		`,
		itemId)

	var saleItemInformation SaleItemInformation
	err := row.Scan(
		&saleItemInformation.Description,
		&saleItemInformation.PriceInCents,
		&saleItemInformation.ItemCategoryId,
		&saleItemInformation.SellCount,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, &ItemNotFoundError{Id: itemId}
	}

	if err != nil {
		return nil, err
	}

	return &saleItemInformation, nil
}
