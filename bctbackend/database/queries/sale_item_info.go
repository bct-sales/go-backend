package queries

import (
	dberr "bctbackend/database/errors"
	models "bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"
)

type SaleItemInformation struct {
	SellerId       models.Id
	Description    string
	ItemCategoryId models.Id
	PriceInCents   models.MoneyInCents
	SellCount      int64
}

// GetSaleItemInformation retrieves information about a sale item.
// If the item is not found, it returns an NoSuchItemError.
func GetSaleItemInformation(
	db *sql.DB,
	itemId models.Id) (*SaleItemInformation, error) {

	row := db.QueryRow(
		`
			SELECT seller_id, description, price_in_cents, item_category_id, COUNT(si.sale_id)
			FROM items i LEFT JOIN sale_items si ON i.item_id = si.item_id
			GROUP BY i.item_id
			HAVING i.item_id = ?
		`,
		itemId)

	var sellerId models.Id
	var description string
	var itemCategoryId models.Id
	var priceInCents models.MoneyInCents
	var sellCount int64
	err := row.Scan(
		&sellerId,
		&description,
		&priceInCents,
		&itemCategoryId,
		&sellCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to get information about item %d: %w", itemId, dberr.ErrNoSuchItem)
	}
	if err != nil {
		return nil, err
	}

	saleItemInformation := SaleItemInformation{
		SellerId:       sellerId,
		Description:    description,
		ItemCategoryId: itemCategoryId,
		PriceInCents:   priceInCents,
		SellCount:      sellCount,
	}

	return &saleItemInformation, nil
}
