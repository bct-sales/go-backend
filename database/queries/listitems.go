package queries

import (
	"bctbackend/database/meta"
	"bctbackend/database/models"
	"bctbackend/database/queries/filters"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

type ListItemsQuery struct {
	filters.LargeItem
	filters.HiddenItem
	filters.Category

	whereClauses *filters.WhereClauses
}

func ListItems() *ListItemsQuery {
	whereClauses := filters.NewWhereClauses()

	query := ListItemsQuery{
		LargeItem:    filters.NewLargeItemFilter(whereClauses),
		HiddenItem:   filters.NewHiddenItemFilter(whereClauses),
		Category:     filters.NewCategoryFilter(whereClauses),
		whereClauses: whereClauses,
	}

	return &query
}

func (query *ListItemsQuery) Execute(db DatabaseQuerier, receiver func(*models.Item) error) (r_err error) {
	sqlQuery := sq.Select(
		meta.Item.ItemID,
		meta.Item.AddedAt,
		meta.Item.Description,
		meta.Item.PriceInCents,
		meta.Item.ItemCategoryID,
		meta.Item.SellerID,
		meta.Item.Donation,
		meta.Item.Charity,
		meta.Item.Frozen,
		meta.Item.Hidden,
		meta.Item.Large,
	).From(meta.Item.Table)

	for _, whereClause := range query.whereClauses.Clauses {
		sqlQuery = sqlQuery.Where(whereClause)
	}

	rows, queryErr := sqlQuery.RunWith(db).Query()
	if queryErr != nil {
		return fmt.Errorf("failed to list items in database: %w", queryErr)
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	// Iterate over rows and call receiver function for each item
	for rows.Next() {
		var itemID models.ID
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var itemCategoryID models.ID
		var sellerID models.ID
		var donation bool
		var charity bool
		var frozen bool
		var hidden bool
		var large bool

		scanErr := rows.Scan(
			&itemID,
			&addedAt,
			&description,
			&priceInCents,
			&itemCategoryID,
			&sellerID,
			&donation,
			&charity,
			&frozen,
			&hidden,
			&large,
		)
		if scanErr != nil {
			return fmt.Errorf("failed to scan row: %w", queryErr)
		}

		item := models.Item{
			ItemID:       itemID,
			AddedAt:      addedAt,
			Description:  description,
			PriceInCents: priceInCents,
			CategoryID:   itemCategoryID,
			SellerID:     sellerID,
			Donation:     donation,
			Charity:      charity,
			Frozen:       frozen,
			Hidden:       hidden,
			Large:        large,
		}

		// If receiver returns error, abort enumeration
		if err := receiver(&item); err != nil {
			return fmt.Errorf("receiver failed: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return nil
}
