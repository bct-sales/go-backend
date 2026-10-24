package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
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

func (q *GetItemsQuery) Execute(db DatabaseQuerier, receiver func(*models.Item) error) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	queryString, queryArguments, err := q.buildSQLQuery()
	if err != nil {
		return err
	}

	// Perform query
	rows, err := db.Query(queryString, queryArguments...)
	if err != nil {
		return fmt.Errorf("failed to execute query %s to look up items in database: %w", queryString, err)
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
		err = rows.Scan(
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
		if err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
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

func (q *GetItemsQuery) buildSQLQuery() (string, []any, error) {
	query := sq.Select("item_id", "added_at", "description", "price_in_cents", "item_category_id", "seller_id", "donation", "charity", "frozen", "hidden", "large")
	query = query.From("items")
	query = query.OrderBy("item_id ASC")

	if q.frozen != nil {
		query = query.Where(sq.Eq{"frozen": *q.frozen})
	}

	if q.hidden != nil {
		query = query.Where(sq.Eq{"hidden": *q.hidden})
	}

	if q.RowRange.Limit != nil {
		query = query.Limit(*q.RowRange.Limit)
	}

	if q.RowRange.Offset != nil {
		// Offset without limit is not allowed
		if q.RowRange.Limit == nil {
			query = query.Limit(100000)
		}

		query = query.Offset(*q.RowRange.Offset)
	}

	if q.categoryID != nil {
		query = query.Where(sq.Eq{"item_category_id": (*q.categoryID).Int64()})
	}

	if q.descriptionPattern != nil {
		query = query.Where(sq.Like{"description": q.descriptionPattern})
	}

	if q.large != nil {
		query = query.Where(sq.Eq{"large": *q.large})
	}

	if q.itemIDs != nil {
		query = query.Where(sq.Eq{"item_id": q.itemIDs})
	}

	queryString, queryArguments, err := query.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	return queryString, queryArguments, nil
}
