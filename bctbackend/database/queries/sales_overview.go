package queries

import (
	models "bctbackend/database/models"
	"database/sql"
	"errors"
)

type CategorySaleTotal struct {
	CategoryId   models.Id
	CategoryName string
	TotalInCents models.MoneyInCents
}

func GetSalesOverview(db *sql.DB) (r_result []CategorySaleTotal, r_err error) {
	rows, err := db.Query(
		`
			SELECT item_categories.item_category_id, item_categories.name, SUM(COALESCE(i.price_in_cents, 0))
			FROM item_categories
			LEFT JOIN (
				items INNER JOIN sale_items ON items.item_id = sale_items.item_id
			) AS i ON i.item_category_id = item_categories.item_category_id
			GROUP BY item_categories.item_category_id
			ORDER BY item_categories.item_category_id
		`,
	)

	if err != nil {
		return nil, err
	}

	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	var categorySaleTotals []CategorySaleTotal

	for rows.Next() {
		var categoryId models.Id
		var categoryName string
		var totalInCents models.MoneyInCents
		err := rows.Scan(
			&categoryId,
			&categoryName,
			&totalInCents,
		)
		if err != nil {
			return nil, err
		}

		categorySaleTotal := CategorySaleTotal{
			CategoryId:   categoryId,
			CategoryName: categoryName,
			TotalInCents: totalInCents,
		}

		categorySaleTotals = append(categorySaleTotals, categorySaleTotal)
	}

	return categorySaleTotals, nil
}
