package queries

import (
	models "bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"
)

type MultiplySoldItem struct {
	Item  models.Item
	Sales []models.Sale
}

func GetMultiplySoldItems(db *sql.DB) (r_result []MultiplySoldItem, r_err error) {
	rows, err := db.Query(
		`
			SELECT item.item_id,
				   item.added_at,
			       item.description,
				   item.price_in_cents,
				   item.item_category_id,
				   item.seller_id,
				   item.donation,
				   item.charity,
				   item.frozen,
				   sale.sale_id,
				   sale.cashier_id,
				   sale.transaction_time
			FROM items item
			INNER JOIN item_categories category ON item.item_category_id = category.item_category_id
			INNER JOIN sale_items sale_item ON item.item_id = sale_item.item_id
			INNER JOIN sales sale ON sale_item.sale_id = sale.sale_id
			WHERE (SELECT COUNT(*)
			       FROM sale_items si
				   WHERE si.item_id = item.item_id) > 1
			ORDER BY item.item_id, sale.sale_id
		`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	var multiplySoldItems []MultiplySoldItem

	for rows.Next() {
		var rowData struct {
			ItemId          models.Id
			AddedAt         models.Timestamp
			Description     string
			PriceInCents    models.MoneyInCents
			CategoryId      models.Id
			SellerId        models.Id
			Donation        bool
			Charity         bool
			Frozen          bool
			SaleId          models.Id
			CashierId       models.Id
			TransactionTime models.Timestamp
		}
		err := rows.Scan(
			&rowData.ItemId,
			&rowData.AddedAt,
			&rowData.Description,
			&rowData.PriceInCents,
			&rowData.CategoryId,
			&rowData.SellerId,
			&rowData.Donation,
			&rowData.Charity,
			&rowData.Frozen,
			&rowData.SaleId,
			&rowData.CashierId,
			&rowData.TransactionTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		sale := models.Sale{
			SaleID:          rowData.SaleId,
			CashierID:       rowData.CashierId,
			TransactionTime: rowData.TransactionTime,
		}

		lastMultiplySoldItemIndex := len(multiplySoldItems) - 1

		if lastMultiplySoldItemIndex >= 0 && multiplySoldItems[lastMultiplySoldItemIndex].Item.ItemID == rowData.ItemId {
			multiplySoldItems[lastMultiplySoldItemIndex].Sales = append(multiplySoldItems[lastMultiplySoldItemIndex].Sales, sale)
		} else {
			multiplySoldItem := MultiplySoldItem{
				Item: models.Item{
					ItemID:       rowData.ItemId,
					AddedAt:      rowData.AddedAt,
					Description:  rowData.Description,
					PriceInCents: rowData.PriceInCents,
					CategoryID:   rowData.CategoryId,
					SellerID:     rowData.SellerId,
					Donation:     rowData.Donation,
					Charity:      rowData.Charity,
					Frozen:       rowData.Frozen,
					Hidden:       false,
				},
				Sales: []models.Sale{sale},
			}

			multiplySoldItems = append(multiplySoldItems, multiplySoldItem)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return multiplySoldItems, nil
}
