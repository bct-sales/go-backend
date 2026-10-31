package queries

import (
	"context"
	"bctbackend/database/meta"
	"bctbackend/database/models"
	"bctbackend/database/queries/filters"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

type (
	ListItemsQuery struct {
		filters.LargeItem
		filters.HiddenItem
		filters.Category

		whereClauses *filters.WhereClauses
	}

	ListItemsQueryRow struct {
		models.Item
		SaleCount int
	}
)

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

func (query *ListItemsQuery) Execute(ctx context.Context, db DatabaseQuerier, receiver func(*ListItemsQueryRow) error) (r_err error) {
	sqlQuery := sq.Select(
		fmt.Sprintf("%s.%s", meta.Item.Table, meta.Item.ItemID),
		fmt.Sprintf("%s.%s", meta.Item.Table, meta.Item.AddedAt),
		fmt.Sprintf("%s.%s", meta.Item.Table, meta.Item.Description),
		fmt.Sprintf("%s.%s", meta.Item.Table, meta.Item.PriceInCents),
		fmt.Sprintf("%s.%s", meta.Item.Table, meta.Item.ItemCategoryID),
		fmt.Sprintf("%s.%s", meta.Item.Table, meta.Item.SellerID),
		fmt.Sprintf("%s.%s", meta.Item.Table, meta.Item.Donation),
		fmt.Sprintf("%s.%s", meta.Item.Table, meta.Item.Charity),
		fmt.Sprintf("%s.%s", meta.Item.Table, meta.Item.Frozen),
		fmt.Sprintf("%s.%s", meta.Item.Table, meta.Item.Hidden),
		fmt.Sprintf("%s.%s", meta.Item.Table, meta.Item.Large),
		fmt.Sprintf("COALESCE(COUNT(%s.%s), 0)", meta.SaleItem.Table, meta.SaleItem.SaleID),
	).From(meta.Item.Table).LeftJoin(fmt.Sprintf("%s ON %s.%s = %s.%s", meta.SaleItem.Table, meta.SaleItem.Table, meta.SaleItem.ItemID, meta.Item.Table, meta.Item.ItemID)).GroupBy(fmt.Sprintf("%s.%s", meta.Item.Table, meta.Item.ItemID))

	for _, whereClause := range query.whereClauses.Clauses {
		sqlQuery = sqlQuery.Where(whereClause)
	}

	rows, queryErr := sqlQuery.RunWith(newSquirrelRunner(ctx, db)).QueryContext(ctx)
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
		var saleCount int

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
			&saleCount,
		)
		if scanErr != nil {
			return fmt.Errorf("failed to scan row: %w", queryErr)
		}

		item := ListItemsQueryRow{
			Item: models.Item{
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
			},
			SaleCount: saleCount,
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
