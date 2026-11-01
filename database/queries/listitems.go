package queries

import (
	"context"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries/filters"
	"bctbackend/database/queries/internal/isql"
	"database/sql"
	"fmt"
)

type (
	ListItemsQuery struct {
		filters.LargeItem
		filters.HiddenItem
		filters.Category
	}

	ListItemsQueryRow struct {
		models.Item
		SaleCount int
	}
)

func ListItems() *ListItemsQuery {
	query := ListItemsQuery{}

	return &query
}

func (query *ListItemsQuery) Execute(ctx context.Context, db DatabaseQuerier, receiver func(*ListItemsQueryRow) error) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	params := isql.ListItemsParams{
		Large:          sql.NullBool{},
		Hidden:         sql.NullBool{},
		ItemCategoryID: sql.NullInt64{},
		Description:    sql.NullString{},
	}

	if large := query.LargeItem.IsLarge(); large != nil {
		params.Large = sql.NullBool{Bool: *large, Valid: true}
	}

	if hidden := query.HiddenItem.IsHidden(); hidden != nil {
		params.Hidden = sql.NullBool{Bool: *hidden, Valid: true}
	}

	if categoryID := query.Category.CategoryID(); categoryID != nil {
		params.ItemCategoryID = sql.NullInt64{Int64: categoryID.Int64(), Valid: true}
	}

	if pattern := query.Category.DescriptionPattern(); pattern != nil {
		params.Description = sql.NullString{String: *pattern, Valid: true}
	}

	rows, err := isql.New(db).ListItems(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to list items in database: %w", err)
	}

	// Iterate over rows and call receiver function for each item
	for i := range rows {
		item := ListItemsQueryRow{
			Item: models.Item{
				ItemID:       rows[i].ItemID,
				AddedAt:      models.Timestamp(rows[i].AddedAt),
				Description:  rows[i].Description,
				PriceInCents: models.MoneyInCents(rows[i].PriceInCents),
				CategoryID:   models.ID(rows[i].ItemCategoryID),
				SellerID:     models.ID(rows[i].SellerID),
				Donation:     rows[i].Donation,
				Charity:      rows[i].Charity,
				Frozen:       rows[i].Frozen,
				Hidden:       rows[i].Hidden,
				Large:        rows[i].Large,
			},
			SaleCount: int(rows[i].SaleCount),
		}

		// If receiver returns error, abort enumeration
		if err := receiver(&item); err != nil {
			return fmt.Errorf("receiver failed: %w", err)
		}
	}

	return nil
}
