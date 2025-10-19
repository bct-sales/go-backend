package queries

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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

type GetItemsQuery struct {
	HiddenFilter
	FrozenFilter
	RowRangeSelection
	DescriptionFilter
	CategoryFilter
}

func NewGetItemsQuery() *GetItemsQuery {
	return &GetItemsQuery{
		HiddenFilter:      HiddenFilter{hidden: nil},
		FrozenFilter:      FrozenFilter{frozen: nil},
		RowRangeSelection: RowRangeSelection{RowRange: RowRange{Limit: nil, Offset: nil}},
		DescriptionFilter: DescriptionFilter{descriptionPattern: nil},
		CategoryFilter:    CategoryFilter{categoryID: nil},
	}
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
	defer func() { err = errors.Join(err, rows.Close()) }()

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
	query := sq.Select("item_id", "added_at", "description", "price_in_cents", "item_category_id", "seller_id", "donation", "charity", "frozen", "hidden")
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

	queryString, queryArguments, err := query.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	return queryString, queryArguments, nil
}

// GetItemIDs retrieves the IDs of all items in the database.
func GetItemIDs(db DatabaseQuerier) (r_result []models.ID, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	// Build SQL query
	query := `
		SELECT
			item_id
		FROM
			items
		ORDER BY
			item_id ASC
	`

	// Perform query
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query to look up item ids in database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	// Iterate over rows and collect item ids
	var itemIDs []models.ID
	for rows.Next() {
		var itemID models.ID
		err = rows.Scan(&itemID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		itemIDs = append(itemIDs, itemID)
	}

	return itemIDs, nil
}

// Returns the items associated with the given seller.
// The itemSelection parameter allows specifying whether to include visible/hidden items or not.
// The items are ordered by their time of addition, then by id.
// An ErrNoSuchUser is returned if no user with the given sellerID exists.
// An ErrWrongRole is returned if sellerID does not refer to a seller.
func GetSellerItems(db DatabaseQuerier, sellerID models.ID, itemSelection ItemSelection) (r_items []*models.Item, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	// Note: GetSellerItems performs multiple queries, but no transaction is necessary
	// since once a user exists with a certain role, it will not disappear.
	if err := EnsureUserExistsAndHasRole(db, sellerID, models.NewSellerRoleID()); err != nil {
		return nil, err
	}

	// Build SQL query
	query := fmt.Sprintf(`
		SELECT
			item_id,
			added_at,
			description,
			price_in_cents,
			item_category_id,
			seller_id,
			donation,
			charity,
			frozen,
			hidden
		FROM
			%s
		WHERE
			seller_id = ?
		ORDER BY
			added_at, item_id ASC
	`, ItemsTableFor(itemSelection))

	rows, err := db.Query(query, sellerID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query to get seller item data from database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	items := make([]*models.Item, 0)

	for rows.Next() {
		var id models.ID
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var itemCategoryID models.ID
		var sellerID models.ID
		var donation bool
		var charity bool
		var frozen bool
		var hidden bool

		err = rows.Scan(
			&id,
			&addedAt,
			&description,
			&priceInCents,
			&itemCategoryID,
			&sellerID,
			&donation,
			&charity,
			&frozen,
			&hidden,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		item := models.Item{
			ItemID:       id,
			AddedAt:      addedAt,
			Description:  description,
			PriceInCents: priceInCents,
			CategoryID:   itemCategoryID,
			SellerID:     sellerID,
			Donation:     donation,
			Charity:      charity,
			Frozen:       frozen,
			Hidden:       hidden,
		}
		items = append(items, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return items, nil
}

type ItemWithSaleCount struct {
	models.Item
	SaleCount int
}

// GetItemsWithSaleCounts looks up all items and how often they have been sold.
// Specifying a non-nil sellerID will return only items from that seller, otherwise items from all sellers are returned.
// The items are ordered by their time of addition, then by id.
// An ErrNoSuchUser is returned if no user with the given sellerID exists.
// An ErrWrongRole is returned if sellerID does not refer to a seller.
func GetItemsWithSaleCounts(db DatabaseQuerier, itemSelection ItemSelection, sellerID *models.ID) (r_items []*ItemWithSaleCount, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	// Note: GetSellerItems performs multiple queries, but no transaction is necessary
	// since once a user exists with a certain role, it will not disappear.
	if sellerID != nil {
		if err := EnsureUserExistsAndHasRole(db, *sellerID, models.NewSellerRoleID()); err != nil {
			return nil, err
		}
	}

	itemsTable := ItemsTableFor(itemSelection)
	var whereClause string
	var arguments []any
	if sellerID != nil {
		whereClause = "WHERE seller_id = ?"
		arguments = append(arguments, *sellerID)
	} else {
		whereClause = ""
	}
	query := fmt.Sprintf(`
		SELECT
			i.item_id,
			added_at,
			description,
			price_in_cents,
			item_category_id,
			seller_id,
			donation,
			charity,
			frozen,
			hidden,
			COALESCE(COUNT(sale_items.sale_id), 0) AS sale_count
		FROM
			%s i LEFT JOIN sale_items ON i.item_id = sale_items.item_id
		%s
		GROUP BY
			i.item_id
		ORDER BY
			added_at, i.item_id ASC
	`, itemsTable, whereClause)
	rows, err := db.Query(query, arguments...)

	if err != nil {
		return nil, fmt.Errorf("failed to execute query to get seller items with sale counts from database: %w", err)
	}

	defer func() { err = errors.Join(err, rows.Close()) }()

	itemsWithSaleCount := make([]*ItemWithSaleCount, 0)

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
		var saleCount int

		err = rows.Scan(&itemID, &addedAt, &description, &priceInCents, &itemCategoryID, &sellerID, &donation, &charity, &frozen, &hidden, &saleCount)
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		itemWithSaleCount := ItemWithSaleCount{
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
			},
			SaleCount: saleCount,
		}

		itemsWithSaleCount = append(itemsWithSaleCount, &itemWithSaleCount)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return itemsWithSaleCount, nil
}

// Returns the items associated with the given seller.
// The items are ordered by their time of addition, then by id.
// Hidden items are not included, as they cannot be sold.
// An ErrNoSuchUser is returned if no user with the given sellerID exists.
// An ErrWrongRole is returned if sellerID does not refer to a seller.
func GetSellerItemsWithSaleCounts(db DatabaseQuerier, sellerID models.ID) (r_items []*ItemWithSaleCount, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	// Note: GetSellerItems performs multiple queries, but no transaction is necessary
	// since once a user exists with a certain role, it will not disappear.
	if err := EnsureUserExistsAndHasRole(db, sellerID, models.NewSellerRoleID()); err != nil {
		return nil, err
	}

	rows, err := db.Query(
		`
			SELECT
				items.item_id,
				added_at,
				description,
				price_in_cents,
				item_category_id,
				seller_id,
				donation,
				charity,
				frozen,
				hidden,
				COALESCE(COUNT(sale_items.sale_id), 0) AS sale_count
			FROM
				items LEFT JOIN sale_items ON items.item_id = sale_items.item_id
			WHERE
				seller_id = ? AND hidden = false
			GROUP BY
				items.item_id
			ORDER BY
				added_at, items.item_id ASC
		`,
		sellerID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query to get seller items with sale counts from database: %w", err)
	}

	defer func() { err = errors.Join(err, rows.Close()) }()

	itemsWithSaleCount := make([]*ItemWithSaleCount, 0)

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
		var saleCount int

		err = rows.Scan(&itemID,
			&addedAt,
			&description,
			&priceInCents,
			&itemCategoryID,
			&sellerID,
			&donation,
			&charity,
			&frozen,
			&hidden,
			&saleCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		itemWithSaleCount := ItemWithSaleCount{
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
			},
			SaleCount: saleCount,
		}

		itemsWithSaleCount = append(itemsWithSaleCount, &itemWithSaleCount)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return itemsWithSaleCount, nil
}

// Returns the item with the given identifier.
// A ErrNoSuchItem is returned if no item with the given identifier exists.
func GetItemWithID(db DatabaseQuerier, itemID models.ID) (r_result *models.Item, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	row := db.QueryRow(`
		SELECT
			added_at,
			description,
			price_in_cents,
			item_category_id,
			seller_id,
			donation,
			charity,
			frozen,
			hidden
		FROM
			items
		WHERE
			item_id = ?
	`, itemID)

	var addedAt models.Timestamp
	var description string
	var priceInCents models.MoneyInCents
	var categoryID models.ID
	var sellerID models.ID
	var donation bool
	var charity bool
	var frozen bool
	var hidden bool
	err := row.Scan(
		&addedAt,
		&description,
		&priceInCents,
		&categoryID,
		&sellerID,
		&donation,
		&charity,
		&frozen,
		&hidden,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dberr.ErrNoSuchItem
		}
		return nil, fmt.Errorf("failed to read row: %w", err)
	}

	item := models.Item{
		ItemID:       itemID,
		AddedAt:      addedAt,
		Description:  description,
		PriceInCents: priceInCents,
		CategoryID:   categoryID,
		SellerID:     sellerID,
		Donation:     donation,
		Charity:      charity,
		Frozen:       frozen,
		Hidden:       hidden,
	}
	return &item, nil
}

// GetItemsWithIDs looks up items with the given IDs.
// The result is a map that relates item IDs to the corresponding item.
// Duplicates in itemIDs are ignored.
// If itemIDs contains a nonexistent item id, a ErrNoSuchItem is returned.
func GetItemsWithIDs(db DatabaseQuerier, itemIDs []models.ID) (r_result map[models.ID]*models.Item, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	// Set up SQL query
	// Note that this does not detect nonexistent items, we deal with that later
	query := fmt.Sprintf(`
		SELECT
			item_id,
			added_at,
			description,
			price_in_cents,
			item_category_id,
			seller_id,
			donation,
			charity,
			frozen,
			hidden
		FROM
			items
		WHERE
			item_id IN (%s)
	`, placeholderString(len(itemIDs)))
	convertedItemIDs := algorithms.Map(itemIDs, func(id models.ID) any { return id })
	rows, err := db.Query(query, convertedItemIDs...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query to get items from database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	items := make(map[models.ID]*models.Item)
	for rows.Next() {
		var id models.ID
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var itemCategoryID models.ID
		var sellerID models.ID
		var donation bool
		var charity bool
		var frozen bool
		var hidden bool

		err = rows.Scan(
			&id,
			&addedAt,
			&description,
			&priceInCents,
			&itemCategoryID,
			&sellerID,
			&donation,
			&charity,
			&frozen,
			&hidden,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		item := models.Item{
			ItemID:       id,
			AddedAt:      addedAt,
			Description:  description,
			PriceInCents: priceInCents,
			CategoryID:   itemCategoryID,
			SellerID:     sellerID,
			Donation:     donation,
			Charity:      charity,
			Frozen:       frozen,
			Hidden:       hidden,
		}
		items[id] = &item
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	// Check if all requested items were found
	if len(items) != len(itemIDs) {
		for _, itemID := range itemIDs {
			if _, ok := items[itemID]; !ok {
				return nil, fmt.Errorf("while getting items, among which %d: %w", itemID, dberr.ErrNoSuchItem)
			}
		}

		// If we get past the loop, it means that all items were found
		// There were duplicates in the requested IDs, but this is not an error
	}

	return items, nil
}

type ItemStatisticsResult struct {
	ItemCount         int
	TotalValueInCents models.MoneyInCents
}

type GetItemStatisticsQuery struct {
	HiddenFilter
	DescriptionFilter
	CategoryFilter
}

func NewGetItemStatisticsQuery() *GetItemStatisticsQuery {
	query := GetItemStatisticsQuery{
		HiddenFilter:      HiddenFilter{hidden: nil},
		DescriptionFilter: DescriptionFilter{descriptionPattern: nil},
		CategoryFilter:    CategoryFilter{categoryID: nil},
	}

	return &query
}

func (q *GetItemStatisticsQuery) Execute(db DatabaseQuerier) (r_result *ItemStatisticsResult, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	query, queryArguments, err := q.buildSQLQuery()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}
	row := db.QueryRow(query, queryArguments...)

	var itemCount int
	var totalValueInCents models.MoneyInCents
	if err := row.Scan(&itemCount, &totalValueInCents); err != nil {
		return nil, fmt.Errorf("failed to read row: %w", err)
	}

	result := ItemStatisticsResult{
		ItemCount:         itemCount,
		TotalValueInCents: totalValueInCents,
	}
	return &result, nil
}

func (q *GetItemStatisticsQuery) buildSQLQuery() (string, []any, error) {
	query := sq.Select("COUNT(item_id)", "COALESCE(SUM(price_in_cents), 0)").From("items")

	if q.hidden != nil {
		query = query.Where(sq.Eq{"hidden": *q.hidden})
	}

	if q.descriptionPattern != nil {
		query = query.Where(sq.Like{"description": fmt.Sprintf("%%%s%%", *q.descriptionPattern)})
	}

	if q.categoryID != nil {
		query = query.Where(sq.Eq{"item_category_id": *q.categoryID})
	}

	queryString, queryArguments, err := query.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	return queryString, queryArguments, nil
}

// GetItemStatistics returns the number of items in the database and their total worth.
// The itemSelection parameter allows specifying which items to count: only hidden, only visible or both.
func GetItemStatistics(db DatabaseQuerier, itemSelection ItemSelection) (r_result *ItemStatisticsResult, r_err error) {
	query := NewGetItemStatisticsQuery()

	switch itemSelection {
	case OnlyHiddenItems:
		query.WithHidden(true)
	case OnlyVisibleItems:
		query.WithHidden(false)
	}

	return query.Execute(db)
}

// AddItem adds an item to the database.
// The ID of the newly added item is returned.
// An ErrNoSuchUser is returned if no user with the given sellerID exists.
// An ErrWrongRole is returned if sellerID does not refer to a seller.
// An ErrNoSuchCategory is returned if the itemCategoryID is invalid.
// An ErrInvalidPrice is returned if the priceInCents is invalid.
// An ErrInvalidItemDescription is returned if the description is invalid.
func AddItem(
	db DatabaseQuerier,
	addedAt models.Timestamp,
	description string,
	priceInCents models.MoneyInCents,
	itemCategoryID models.ID,
	sellerID models.ID,
	donation bool,
	charity bool,
	frozen bool,
	hidden bool) (r_result models.ID, r_err error) {

	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	// Validate inputs
	if !models.IsValidPrice(priceInCents) {
		return 0, fmt.Errorf("failed to add item with price %d: %w", priceInCents, dberr.ErrInvalidPrice)
	}
	if !models.IsValidItemDescription(description) {
		return 0, fmt.Errorf("failed to add item with description %s: %w", description, dberr.ErrInvalidItemDescription)
	}
	// No transaction is necessary here, since users don't change
	if err := EnsureUserExistsAndHasRole(db, sellerID, models.NewSellerRoleID()); err != nil {
		return 0, fmt.Errorf("could not ensure user %d exists and is seller: %w", sellerID, err)
	}
	if frozen && hidden {
		return 0, fmt.Errorf("failed to add item: %w", dberr.ErrHiddenFrozenItem)
	}

	// Insert the item into the database
	result, err := db.Exec(
		`
			INSERT INTO items (added_at, description, price_in_cents, item_category_id, seller_id, donation, charity, frozen, hidden)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`,
		addedAt,
		description,
		priceInCents,
		itemCategoryID,
		sellerID,
		donation,
		charity,
		frozen,
		hidden,
	)
	if err != nil {
		categoryExists, err2 := CategoryWithIDExists(db, itemCategoryID)
		if err2 != nil {
			return 0, fmt.Errorf("failed to determine whether category with given id exists: %w", err)
		}

		if !categoryExists {
			return 0, fmt.Errorf("failed to add item with category %d: %w", itemCategoryID, dberr.ErrNoSuchCategory)
		}

		return 0, fmt.Errorf("failed to insert item: %w", err)
	}

	// Get ID of the inserted item
	itemID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to determine id of inserted item: %w", err)
	}

	return models.ID(itemID), nil
}

// ItemWithIDExists returns true if an item with the given identifier exists in the database.
func ItemWithIDExists(db DatabaseQuerier, itemID models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	row := db.QueryRow(
		`
			SELECT
				1
			FROM
				items
			WHERE
				item_id = $1
		`,
		itemID,
	)

	var result int
	err := row.Scan(&result)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

// ItemsExists checks if all given items exist in the database.
// Duplicates in itemIDs have no effect on the result.
// Returns true if all items exist, false otherwise.
func ItemsExist(db DatabaseQuerier, itemIDs []models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if len(itemIDs) == 0 {
		return true, nil
	}

	itemIDs = algorithms.RemoveDuplicates(itemIDs)

	// Set up SQL query
	query := fmt.Sprintf(`
		SELECT
			COUNT(item_id)
		FROM
			items
		WHERE
			item_id IN (%s)
	`, placeholderString(len(itemIDs)))

	convertedItemIDs := algorithms.Map(itemIDs, func(id models.ID) any { return id })
	row := db.QueryRow(query, convertedItemIDs...)

	var count int
	err := row.Scan(&count)

	if err != nil {
		return false, err
	}

	return count == len(itemIDs), nil
}

// EnsureItemsExist checks if all items with the given IDs exist in the database.
// If any item does not exist, it returns a ErrNoSuchItem.
func EnsureItemsExist(db DatabaseQuerier, itemIDs []models.ID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	itemsExist, err := ItemsExist(db, itemIDs)
	if err != nil {
		return fmt.Errorf("failed to ensure items exist: %w", err)
	}
	if !itemsExist {
		return fmt.Errorf("failed to ensure items exist: %w", dberr.ErrNoSuchItem)
	}

	return nil
}

// UpdateFreezeStatusOfItems updates the frozen status of multiple items at once.
// If any item does not exist, it returns an ErrNoSuchItem.
// If any item is hidden, it returns a ErrItemHidden error.
// Duplicates in itemIDs are ignored.
// In case of an error, no items are updated.
// This function consists of multiple database interactions, so it must run within a transaction.
func UpdateFreezeStatusOfItems(transaction *TransactionalDatabaseQuerier, itemIDs []models.ID, frozen bool) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if len(itemIDs) == 0 {
		return nil
	}

	itemIDs = algorithms.RemoveDuplicates(itemIDs)
	convertedItemIDs := algorithms.Map(itemIDs, func(id models.ID) any { return id })

	if err := EnsureItemsExist(transaction, itemIDs); err != nil {
		return err
	}

	if err := EnsureNoHiddenItems(transaction, itemIDs); err != nil {
		return fmt.Errorf("failed to ensure no hidden items: %w", err)
	}

	query := fmt.Sprintf(`
		UPDATE items
		SET frozen = ?
		WHERE item_id IN (%s)
	`, placeholderString(len(itemIDs)))

	sqlValues := append([]any{frozen}, convertedItemIDs...)

	if _, err := transaction.Exec(query, sqlValues...); err != nil {
		return err
	}

	return nil
}

// UpdateHiddenStatusOfItems updates the hidden status of multiple items at once.
// If any item does not exist, it returns an ErrNoSuchItem.
// If any item is hidden, it returns a ErrItemFrozen error.
// Duplicates in itemIDs are ignored.
// In case of an error, no items are updated.
// This function consists of multiple database interactions, so it must run within a transaction.
func UpdateHiddenStatusOfItems(transaction *TransactionalDatabaseQuerier, itemIDs []models.ID, hidden bool) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if len(itemIDs) == 0 {
		return nil
	}

	itemIDs = algorithms.RemoveDuplicates(itemIDs)
	convertedItemIDs := algorithms.Map(itemIDs, func(id models.ID) any { return id })

	if err := EnsureItemsExist(transaction, itemIDs); err != nil {
		return err
	}

	// Check if none of the items are frozen
	if err := EnsureNoFrozenItems(transaction, itemIDs); err != nil {
		return err
	}

	query := fmt.Sprintf(`
		UPDATE items
		SET hidden = ?
		WHERE item_id IN (%s)
	`, placeholderString(len(itemIDs)))

	sqlValues := append([]any{hidden}, convertedItemIDs...)

	if _, err := transaction.Exec(query, sqlValues...); err != nil {
		return err
	}

	return nil
}

func partitionItemsBy(database DatabaseQuerier, itemIDs []models.ID, columnName string) (*algorithms.Set[models.ID], *algorithms.Set[models.ID], error) {
	query := fmt.Sprintf(`
		SELECT
			item_id,
			%s
		FROM
			items
		WHERE
			item_id IN (%s)
	`, columnName, placeholderString(len(itemIDs)))
	convertedItemIDs := algorithms.Map(itemIDs, func(id models.ID) any { return id })
	rows, err := database.Query(query, convertedItemIDs...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()

	falseSet := algorithms.NewSet[models.ID]()
	trueSet := algorithms.NewSet[models.ID]()
	for rows.Next() {
		var id models.ID
		var hiddenStatus bool

		err = rows.Scan(&id, &hiddenStatus)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan items: %w", err)
		}

		if hiddenStatus {
			trueSet.Add(id)
		} else {
			falseSet.Add(id)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return &falseSet, &trueSet, nil
}

// PartitionItemsByHiddenStatus partitions the given item IDs into two sets: one for unhidden items and one for hidden items.
// If an item ID does not exist in the database, it is ignored.
func PartitionItemsByHiddenStatus(db DatabaseQuerier, itemIDs []models.ID) (r_visible *algorithms.Set[models.ID], r_hidden *algorithms.Set[models.ID], r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	return partitionItemsBy(db, itemIDs, "hidden")
}

// PartitionItemsByFrozenStatus partitions the given item IDs into two sets: one for nonfrozen items and one for frozen items.
// If an item ID does not exist in the database, it is ignored.
func PartitionItemsByFrozenStatus(db DatabaseQuerier, itemIDs []models.ID) (r_nonfrozen *algorithms.Set[models.ID], r_frozen *algorithms.Set[models.ID], r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	return partitionItemsBy(db, itemIDs, "frozen")
}

// ContainsHiddenItems checks if any of the given items are hidden.
// It returns true if at least one item is hidden, and false otherwise.
// It is not an error when nonexistent items are passed in, they are simply ignored.
func ContainsHiddenItems(database DatabaseQuerier, itemIDs []models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if len(itemIDs) == 0 {
		return false, nil
	}

	_, hidden, err := PartitionItemsByHiddenStatus(database, itemIDs)
	if err != nil {
		return false, err
	}

	containsHidden := hidden.Len() > 0

	return containsHidden, nil
}

// ContainsFrozenItems checks if any of the given items are frozen.
// It returns true if at least one item is frozen, and false otherwise.
// It is not an error when nonexistent items are passed in, they are simply ignored.
func ContainsFrozenItems(database DatabaseQuerier, itemIDs []models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if len(itemIDs) == 0 {
		return false, nil
	}

	_, frozen, err := PartitionItemsByFrozenStatus(database, itemIDs)
	if err != nil {
		return false, err
	}

	containsFrozen := frozen.Len() > 0

	return containsFrozen, nil
}

// IsItemFrozen checks if none of the items is frozen.
// If one or more items are frozen, it returns an ErrItemFrozen error.
func EnsureNoFrozenItems(qh DatabaseQuerier, itemIDs []models.ID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	containsFrozen, err := ContainsFrozenItems(qh, itemIDs)

	if err != nil {
		return fmt.Errorf("failed to check for frozen items: %w", err)
	}

	if containsFrozen {
		return dberr.ErrItemFrozen
	}

	return nil
}

// EnsureNoHiddenItems checks if none of the items is hidden.
// If one or more items are hidden, it returns an ErrItemHidden error.
func EnsureNoHiddenItems(qh DatabaseQuerier, itemIDs []models.ID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	containsHidden, err := ContainsHiddenItems(qh, itemIDs)

	if err != nil {
		return fmt.Errorf("failed to check for hidden items: %w", err)
	}

	if containsHidden {
		return dberr.ErrItemHidden
	}

	return nil
}

// IsItemFrozen checks whether the item with the given ID is frozen.
// ErrNoSuchItem is returned if the item does not exist.
func IsItemFrozen(db DatabaseQuerier, itemID models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	nonFrozen, frozen, err := PartitionItemsByFrozenStatus(db, []models.ID{itemID})
	if err != nil {
		return false, err
	}

	isFrozen := frozen.Len() > 0
	if isFrozen {
		return true, nil
	}

	isUnfrozen := nonFrozen.Len() > 0
	if isUnfrozen {
		return false, nil
	}

	return false, fmt.Errorf("failed to check if item %d is frozen: %w", itemID, dberr.ErrNoSuchItem)
}

// IsItemHidden checks whether the item with the given ID is hidden.
// ErrNoSuchItem is returned if the item does not exist.
func IsItemHidden(db DatabaseQuerier, itemID models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	unhidden, hidden, err := PartitionItemsByHiddenStatus(db, []models.ID{itemID})
	if err != nil {
		return false, err
	}

	isFrozen := hidden.Len() > 0
	if isFrozen {
		return true, nil
	}

	isUnfrozen := unhidden.Len() > 0
	if isUnfrozen {
		return false, nil
	}

	return false, fmt.Errorf("failed to check if item %d is hidden: %w", itemID, dberr.ErrNoSuchItem)
}

// RemoveItemWithID removes the item with the given ID from the database.
// This function breaks the monotonicity invariant upon which other functionality relies,
// so it should be used with caution.
// If the item does not exist, ErrNoSuchItem is returned.
// If the item has been sold, ErrItemSold is returned and the item remains in the database.
// If the item is frozen, ErrItemFrozen is returned and the item remains in the database.
func RemoveItemWithID(db DatabaseQuerier, itemID models.ID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if err := EnsureItemsExist(db, []models.ID{itemID}); err != nil {
		return fmt.Errorf("failed to remove item with id %d: %w", itemID, dberr.ErrNoSuchItem)
	}

	if err := EnsureNoFrozenItems(db, []models.ID{itemID}); err != nil {
		return fmt.Errorf("failed to remove item with id %d: %w", itemID, dberr.ErrItemFrozen)
	}

	_, err := db.Exec(
		`
			DELETE FROM items
			WHERE item_id = $1
		`,
		itemID,
	)
	if err != nil {
		sold, err2 := HasAnyBeenSold(db, []models.ID{itemID})
		if err2 != nil {
			return err
		}
		if sold {
			return fmt.Errorf("failed to remove item with id %d: %w", itemID, dberr.ErrItemSold)
		}

		return fmt.Errorf("failed to remove item with id %d: %w", itemID, err)
	}

	return nil
}

// ItemUpdate represents the fields that can be updated in an item.
// It is used by the function UpdateItem.
type ItemUpdate struct {
	AddedAt      *models.Timestamp    // If nil, the AddedAt field is not updated.
	Description  *string              // If nil, the Description field is not updated.
	PriceInCents *models.MoneyInCents // If nil, the PriceInCents field is not updated.
	CategoryID   *models.ID           // If nil, the CategoryID field is not updated.
	Donation     *bool                // If nil, the Donation field is not updated.
	Charity      *bool                // If nil, the Charity field is not updated.
}

// UpdateItem updates the item with the given ID in the database.
// If the item does not exist, ErrNoSuchItem is returned.
// If the item is frozen, ErrItemFrozen is returned.
// If the item is hidden, ErrItemHidden is returned.
func UpdateItem(db *TransactionalDatabaseQuerier, itemID models.ID, itemUpdate *ItemUpdate) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	item, err := GetItemWithID(db, itemID)
	if err != nil {
		return err
	}

	if item.Frozen {
		return dberr.ErrItemFrozen
	}

	if item.Hidden {
		return dberr.ErrItemHidden
	}

	sqlUpdates := []string{}
	sqlValues := []any{}

	if itemUpdate.AddedAt != nil {
		sqlUpdates = append(sqlUpdates, "added_at = ?")
		sqlValues = append(sqlValues, *itemUpdate.AddedAt)
	}

	if itemUpdate.Description != nil {
		sqlUpdates = append(sqlUpdates, "description = ?")
		sqlValues = append(sqlValues, *itemUpdate.Description)
	}

	if itemUpdate.PriceInCents != nil {
		if !models.IsValidPrice(*itemUpdate.PriceInCents) {
			return fmt.Errorf("failed to updated item's price to %d: %w", *itemUpdate.PriceInCents, dberr.ErrInvalidPrice)
		}

		sqlUpdates = append(sqlUpdates, "price_in_cents = ?")
		sqlValues = append(sqlValues, *itemUpdate.PriceInCents)
	}

	if itemUpdate.CategoryID != nil {
		categoryExists, err := CategoryWithIDExists(db, *itemUpdate.CategoryID)
		if err != nil {
			return err
		}

		if !categoryExists {
			return fmt.Errorf("failed to update item's category to %d", *itemUpdate.CategoryID)
		}

		sqlUpdates = append(sqlUpdates, "item_category_id = ?")
		sqlValues = append(sqlValues, *itemUpdate.CategoryID)
	}

	if itemUpdate.Donation != nil {
		sqlUpdates = append(sqlUpdates, "donation = ?")
		sqlValues = append(sqlValues, *itemUpdate.Donation)
	}

	if itemUpdate.Charity != nil {
		sqlUpdates = append(sqlUpdates, "charity = ?")
		sqlValues = append(sqlValues, *itemUpdate.Charity)
	}

	if len(sqlUpdates) == 0 {
		return nil
	}

	sqlValues = append(sqlValues, itemID)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE item_id = ?", "items", strings.Join(sqlUpdates, ", "))

	if _, err := db.Exec(query, sqlValues...); err != nil {
		return err
	}

	return nil
}

type AddItemFunction func(addedAt models.Timestamp, description string, priceInCents models.MoneyInCents, itemCategoryID models.ID, sellerID models.ID, donation bool, charity bool, frozen bool, hidden bool)

type AddItemsCallback func(addItem AddItemFunction)

// AddItems allows to add multiple items to the database at once.
func AddItems(db DatabaseQuerier, callback AddItemsCallback) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	valuesString := []string{}
	arguments := []any{}
	tupleString := "(?, ?, ?, ?, ?, ?, ?, ?, ?)"

	add := func(addedAt models.Timestamp, description string, priceInCents models.MoneyInCents, itemCategoryID models.ID, sellerID models.ID, donation bool, charity bool, frozen bool, hidden bool) {
		valuesString = append(valuesString, tupleString)
		arguments = append(arguments, addedAt, description, priceInCents, itemCategoryID, sellerID, donation, charity, frozen, hidden)
	}

	callback(add)

	if len(valuesString) == 0 {
		return nil
	}

	query := `INSERT INTO items (added_at, description, price_in_cents, item_category_id, seller_id, donation, charity, frozen, hidden) VALUES ` + strings.Join(valuesString, ",")

	if _, err := db.Exec(query, arguments...); err != nil {
		return err
	}

	return nil
}

// DoesSellerHaveFrozenItems checks if any item owned by the given seller is frozen.
// Returns ErrNoSuchUser is sellerID does not exist.
// Returns ErrWrongRole if sellerID does not refer to a seller.
func DoesSellerHaveFrozenItems(db *TransactionalDatabaseQuerier, sellerID models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if err := EnsureUserExistsAndHasRole(db, sellerID, models.NewSellerRoleID()); err != nil {
		return false, err
	}

	query := `
		SELECT
			1
		FROM
			items
		WHERE
			seller_id = ? AND items.frozen
	`

	row := db.QueryRow(query, sellerID)

	var dummy uint64
	if err := row.Scan(&dummy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// MoveItemsToNewSeller moves all items belonging to oldSellerID to newSellerID.
// Does not check for frozen or hidden items!
// Returns ErrNoSuchUser if oldSellerID or newSellerID do not exist.
// Returns ErrWrongRole if oldSellerID or newSellerID do not refer to sellers.
func MoveItemsToNewSeller(db *TransactionalDatabaseQuerier, oldSellerID models.ID, newSellerID models.ID) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	// oldSellerID must refer to seller
	if err := EnsureUserExistsAndHasRole(db, oldSellerID, models.NewSellerRoleID()); err != nil {
		return err
	}

	// newSellerID must refer to seller
	if err := EnsureUserExistsAndHasRole(db, newSellerID, models.NewSellerRoleID()); err != nil {
		return err
	}

	query := `
		UPDATE items
		SET seller_id = ?
		WHERE seller_id = ?
	`
	if _, err := db.Exec(query, newSellerID, oldSellerID); err != nil {
		return err
	}

	return nil
}
