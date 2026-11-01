package queries

import (
	"context"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries/internal/isql"
	"database/sql"
	"fmt"
)

type HiddenFilter struct {
	hidden *bool
}

func (filter *HiddenFilter) WithHidden(value bool) {
	filter.hidden = &value
}

type FrozenFilter struct {
	frozen *bool
}

func (filter *FrozenFilter) WithFrozen(value bool) {
	filter.frozen = &value
}

type RowRangeSelection struct {
	RowRange
}

func (q *GetItemsQuery) WithRowRange(rowRange *RowRange) {
	q.RowRange = *rowRange
}

type DescriptionFilter struct {
	descriptionPattern *string
}

func (filter *DescriptionFilter) WithDescription(pattern string) {
	filter.descriptionPattern = &pattern
}

type CategoryFilter struct {
	categoryID *models.ID
}

func (filter *CategoryFilter) WithCategory(categoryID models.ID) {
	filter.categoryID = &categoryID
}

type LargeFilter struct {
	large *bool
}

func (filter *LargeFilter) WithLarge(value bool) {
	filter.large = &value
}

type ItemIDSelectionFilter struct {
	itemIDs []models.ID
}

func (filter *ItemIDSelectionFilter) WithItemIDs(itemIDs []models.ID) {
	filter.itemIDs = itemIDs
}

type GetItemsQuery struct {
	HiddenFilter
	FrozenFilter
	RowRangeSelection
	DescriptionFilter
	CategoryFilter
	LargeFilter
	ItemIDSelectionFilter
}

func NewGetItemsQuery() *GetItemsQuery {
	return &GetItemsQuery{}
}

func (q *GetItemsQuery) WithCategory(categoryID models.ID) {
	q.categoryID = &categoryID
}

func (q *GetItemsQuery) WithDescriptionPattern(descriptionSubstring string) {
	q.descriptionPattern = &descriptionSubstring
}

func (q *GetItemsQuery) Execute(ctx context.Context, db DatabaseQuerier, receiver func(*models.Item) error) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	// The generated queries are split in two because sqlc's slice expansion is
	// incompatible with trailing LIMIT/OFFSET parameters (they would share
	// placeholder indices). No caller combines item IDs with a row range.
	// Mirror the previous dynamic query: an unset limit means no LIMIT clause
	// at all (sentinel -1, which SQLite treats as unbounded), and an OFFSET
	// without LIMIT used an implicit 100000-row limit.
	limit := int64(-1)
	offset := int64(0)
	if q.RowRange.Limit != nil {
		limit = int64(*q.RowRange.Limit)
	} else if q.RowRange.Offset != nil {
		limit = 100000
	}
	if q.RowRange.Offset != nil {
		offset = int64(*q.RowRange.Offset)
	}

	var rows []isql.Item
	var err error

	if q.itemIDs != nil {
		params := isql.GetItemsWithIDsParams{
			Hidden:         sql.NullBool{},
			Frozen:         sql.NullBool{},
			ItemCategoryID: sql.NullInt64{},
			Description:    sql.NullString{},
			Large:          sql.NullBool{},
			ItemIds:        q.itemIDs,
		}

		if q.hidden != nil {
			params.Hidden = sql.NullBool{Bool: *q.hidden, Valid: true}
		}

		if q.frozen != nil {
			params.Frozen = sql.NullBool{Bool: *q.frozen, Valid: true}
		}

		if q.categoryID != nil {
			params.ItemCategoryID = sql.NullInt64{Int64: q.categoryID.Int64(), Valid: true}
		}

		if q.descriptionPattern != nil {
			params.Description = sql.NullString{String: *q.descriptionPattern, Valid: true}
		}

		if q.large != nil {
			params.Large = sql.NullBool{Bool: *q.large, Valid: true}
		}

		rows, err = isql.New(db).GetItemsWithIDs(ctx, params)
	} else {
		params := isql.GetItemsParams{
			Hidden:         sql.NullBool{},
			Frozen:         sql.NullBool{},
			ItemCategoryID: sql.NullInt64{},
			Description:    sql.NullString{},
			Large:          sql.NullBool{},
			Limit:          limit,
			Offset:         offset,
		}

		if q.hidden != nil {
			params.Hidden = sql.NullBool{Bool: *q.hidden, Valid: true}
		}

		if q.frozen != nil {
			params.Frozen = sql.NullBool{Bool: *q.frozen, Valid: true}
		}

		if q.categoryID != nil {
			params.ItemCategoryID = sql.NullInt64{Int64: q.categoryID.Int64(), Valid: true}
		}

		if q.descriptionPattern != nil {
			params.Description = sql.NullString{String: *q.descriptionPattern, Valid: true}
		}

		if q.large != nil {
			params.Large = sql.NullBool{Bool: *q.large, Valid: true}
		}

		rows, err = isql.New(db).GetItems(ctx, params)
	}

	if err != nil {
		return fmt.Errorf("failed to execute query to look up items in database: %w", err)
	}

	// Iterate over rows and call receiver function for each item
	for i := range rows {
		item := models.Item{
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
		}

		// If receiver returns error, abort enumeration
		if err := receiver(&item); err != nil {
			return fmt.Errorf("receiver failed: %w", err)
		}
	}

	return nil
}
