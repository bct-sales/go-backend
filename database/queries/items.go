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

type GetItemsQuery struct {
	frozen *bool
	hidden *bool
	limit  *uint64
	offset *uint64
}

func NewGetItemsQuery() *GetItemsQuery {
	return &GetItemsQuery{
		frozen: nil,
		hidden: nil,
		limit:  nil,
		offset: nil,
	}
}

func (q *GetItemsQuery) WithFrozen(value bool) {
	q.frozen = &value
}

func (q *GetItemsQuery) WithHidden(value bool) {
	q.hidden = &value
}

func (q *GetItemsQuery) WithLimitAndOffset(limit uint64, offset uint64) {
	q.limit = &limit
	q.offset = &offset
}

func (q *GetItemsQuery) Execute(db DatabaseQuerier, receiver func(*models.Item) error) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	queryString, queryArguments, err := q.buildSqlQuery()

	// Perform query
	rows, err := db.Query(queryString, queryArguments...)
	if err != nil {
		return fmt.Errorf("failed to execute query %s to look up items in database: %w", queryString, err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()

	// Iterate over rows and call receiver function for each item
	for rows.Next() {
		var itemID models.Id
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var itemCategoryId models.Id
		var sellerID models.Id
		var donation bool
		var charity bool
		var frozen bool
		var hidden bool
		err = rows.Scan(
			&itemID,
			&addedAt,
			&description,
			&priceInCents,
			&itemCategoryId,
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
			CategoryID:   itemCategoryId,
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

func (q *GetItemsQuery) buildSqlQuery() (string, []any, error) {
	query := sq.Select("item_id", "added_at", "description", "price_in_cents", "item_category_id", "seller_id", "donation", "charity", "frozen", "hidden")
	query = query.From("items")
	query = query.OrderBy("item_id ASC")

	if q.frozen != nil {
		query = query.Where(sq.Eq{"frozen": *q.frozen})
	}

	if q.hidden != nil {
		query = query.Where(sq.Eq{"hidden": *q.hidden})
	}

	if q.limit != nil {
		query = query.Limit(*q.limit)
	}

	if q.offset != nil {
		// Offset without limit is not allowed
		if q.limit == nil {
			query = query.Limit(100000)
		}

		query = query.Offset(*q.offset)
	}

	queryString, queryArguments, err := query.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	return queryString, queryArguments, nil
}

// GetItems can be used to fetch items from the database.
// For each item found, the given receiver function is called.
// If an error occurs while processing an item, the error is returned.
// The itemSelection parameter specifies which items to retrieve: only hidden, only visible or both.
// The rowSelection parameter specifies which rows to retrieve.
func GetItems(db DatabaseQuerier, receiver func(*models.Item) error, itemSelection ItemSelection, rowSelection *RowSelection) error {
	query := NewGetItemsQuery()

	switch itemSelection {
	case AllItems:
		// NOP
	case OnlyVisibleItems:
		query.WithHidden(false)
	case OnlyHiddenItems:
		query.WithHidden(true)
	default:
		panic(fmt.Sprintf("Invalid hidden strategy: %d", itemSelection))
	}

	if rowSelection != nil && (rowSelection.Offset != nil || rowSelection.Limit != nil) {
		limit := uint64(100000)
		offset := uint64(0)

		if rowSelection.Limit != nil {
			limit = uint64(*rowSelection.Limit)
		}

		if rowSelection.Offset != nil {
			offset = uint64(*rowSelection.Offset)
		}

		query.WithLimitAndOffset(limit, offset)
	}

	return query.Execute(db, receiver)
}

// GetItemIds retrieves the IDs of all items in the database.
func GetItemIds(db DatabaseQuerier) (r_result []models.Id, r_err error) {
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
	var itemIds []models.Id
	for rows.Next() {
		var itemId models.Id
		err = rows.Scan(&itemId)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		itemIds = append(itemIds, itemId)
	}

	return itemIds, nil
}

// Returns the items associated with the given seller.
// The itemSelection parameter allows specifying whether to include visible/hidden items or not.
// The items are ordered by their time of addition, then by id.
// An ErrNoSuchUser is returned if no user with the given sellerId exists.
// An ErrWrongRole is returned if sellerId does not refer to a seller.
func GetSellerItems(db DatabaseQuerier, sellerId models.Id, itemSelection ItemSelection) (r_items []*models.Item, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	// Note: GetSellerItems performs multiple queries, but no transaction is necessary
	// since once a user exists with a certain role, it will not disappear.
	if err := EnsureUserExistsAndHasRole(db, sellerId, models.NewSellerRoleId()); err != nil {
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

	rows, err := db.Query(query, sellerId)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query to get seller item data from database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	items := make([]*models.Item, 0)

	for rows.Next() {
		var id models.Id
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var itemCategoryId models.Id
		var sellerId models.Id
		var donation bool
		var charity bool
		var frozen bool
		var hidden bool

		err = rows.Scan(
			&id,
			&addedAt,
			&description,
			&priceInCents,
			&itemCategoryId,
			&sellerId,
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
			CategoryID:   itemCategoryId,
			SellerID:     sellerId,
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
// Specifying a non-nil sellerId will return only items from that seller, otherwise items from all sellers are returned.
// The items are ordered by their time of addition, then by id.
// An ErrNoSuchUser is returned if no user with the given sellerId exists.
// An ErrWrongRole is returned if sellerId does not refer to a seller.
func GetItemsWithSaleCounts(db DatabaseQuerier, itemSelection ItemSelection, sellerId *models.Id) (r_items []*ItemWithSaleCount, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	// Note: GetSellerItems performs multiple queries, but no transaction is necessary
	// since once a user exists with a certain role, it will not disappear.
	if sellerId != nil {
		if err := EnsureUserExistsAndHasRole(db, *sellerId, models.NewSellerRoleId()); err != nil {
			return nil, err
		}
	}

	itemsTable := ItemsTableFor(itemSelection)
	var whereClause string
	var arguments []any
	if sellerId != nil {
		whereClause = "WHERE seller_id = ?"
		arguments = append(arguments, *sellerId)
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
		var itemID models.Id
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var itemCategoryID models.Id
		var sellerID models.Id
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
// An ErrNoSuchUser is returned if no user with the given sellerId exists.
// An ErrWrongRole is returned if sellerId does not refer to a seller.
func GetSellerItemsWithSaleCounts(db DatabaseQuerier, sellerId models.Id) (r_items []*ItemWithSaleCount, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	// Note: GetSellerItems performs multiple queries, but no transaction is necessary
	// since once a user exists with a certain role, it will not disappear.
	if err := EnsureUserExistsAndHasRole(db, sellerId, models.NewSellerRoleId()); err != nil {
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
		sellerId,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query to get seller items with sale counts from database: %w", err)
	}

	defer func() { err = errors.Join(err, rows.Close()) }()

	itemsWithSaleCount := make([]*ItemWithSaleCount, 0)

	for rows.Next() {
		var itemID models.Id
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var itemCategoryId models.Id
		var sellerID models.Id
		var donation bool
		var charity bool
		var frozen bool
		var hidden bool
		var saleCount int

		err = rows.Scan(&itemID,
			&addedAt,
			&description,
			&priceInCents,
			&itemCategoryId,
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
				CategoryID:   itemCategoryId,
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
func GetItemWithId(db DatabaseQuerier, itemId models.Id) (r_result *models.Item, r_err error) {
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
	`, itemId)

	var addedAt models.Timestamp
	var description string
	var priceInCents models.MoneyInCents
	var categoryId models.Id
	var sellerId models.Id
	var donation bool
	var charity bool
	var frozen bool
	var hidden bool
	err := row.Scan(
		&addedAt,
		&description,
		&priceInCents,
		&categoryId,
		&sellerId,
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
		ItemID:       itemId,
		AddedAt:      addedAt,
		Description:  description,
		PriceInCents: priceInCents,
		CategoryID:   categoryId,
		SellerID:     sellerId,
		Donation:     donation,
		Charity:      charity,
		Frozen:       frozen,
		Hidden:       hidden,
	}
	return &item, nil
}

// GetItemsWithIds looks up items with the given IDs.
// The result is a map that relates item IDs to the corresponding item.
// Duplicates in itemIds are ignored.
// If itemIds contains a nonexistent item id, a ErrNoSuchItem is returned.
func GetItemsWithIds(db DatabaseQuerier, itemIds []models.Id) (r_result map[models.Id]*models.Item, r_err error) {
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
	`, placeholderString(len(itemIds)))
	convertedItemIds := algorithms.Map(itemIds, func(id models.Id) any { return id })
	rows, err := db.Query(query, convertedItemIds...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query to get items from database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	items := make(map[models.Id]*models.Item)
	for rows.Next() {
		var id models.Id
		var addedAt models.Timestamp
		var description string
		var priceInCents models.MoneyInCents
		var itemCategoryId models.Id
		var sellerId models.Id
		var donation bool
		var charity bool
		var frozen bool
		var hidden bool

		err = rows.Scan(
			&id,
			&addedAt,
			&description,
			&priceInCents,
			&itemCategoryId,
			&sellerId,
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
			CategoryID:   itemCategoryId,
			SellerID:     sellerId,
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
	if len(items) != len(itemIds) {
		for _, itemId := range itemIds {
			if _, ok := items[itemId]; !ok {
				return nil, fmt.Errorf("while getting items, among which %d: %w", itemId, dberr.ErrNoSuchItem)
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

// GetItemStatistics returns the number of items in the database and their total worth.
// The itemSelection parameter allows specifying which items to count: only hidden, only visible or both.
func GetItemStatistics(db DatabaseQuerier, itemSelection ItemSelection) (r_result *ItemStatisticsResult, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	query := fmt.Sprintf(`
		SELECT
			COUNT(item_id), COALESCE(SUM(price_in_cents), 0)
		FROM
			%s
	`, ItemsTableFor(itemSelection))
	row := db.QueryRow(query)

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

// AddItem adds an item to the database.
// The ID of the newly added item is returned.
// An ErrNoSuchUser is returned if no user with the given sellerId exists.
// An ErrWrongRole is returned if sellerId does not refer to a seller.
// An ErrNoSuchCategory is returned if the itemCategoryId is invalid.
// An ErrInvalidPrice is returned if the priceInCents is invalid.
// An ErrInvalidItemDescription is returned if the description is invalid.
func AddItem(
	db DatabaseQuerier,
	addedAt models.Timestamp,
	description string,
	priceInCents models.MoneyInCents,
	itemCategoryId models.Id,
	sellerId models.Id,
	donation bool,
	charity bool,
	frozen bool,
	hidden bool) (r_result models.Id, r_err error) {

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
	if err := EnsureUserExistsAndHasRole(db, sellerId, models.NewSellerRoleId()); err != nil {
		return 0, fmt.Errorf("could not ensure user %d exists and is seller: %w", sellerId, err)
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
		itemCategoryId,
		sellerId,
		donation,
		charity,
		frozen,
		hidden,
	)
	if err != nil {
		categoryExists, err2 := CategoryWithIdExists(db, itemCategoryId)
		if err2 != nil {
			return 0, fmt.Errorf("failed to determine whether category with given id exists: %w", err)
		}

		if !categoryExists {
			return 0, fmt.Errorf("failed to add item with category %d: %w", itemCategoryId, dberr.ErrNoSuchCategory)
		}

		return 0, fmt.Errorf("failed to insert item: %w", err)
	}

	// Get ID of the inserted item
	itemId, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to determine id of inserted item: %w", err)
	}

	return models.Id(itemId), nil
}

// ItemWithIdExists returns true if an item with the given identifier exists in the database.
func ItemWithIdExists(db DatabaseQuerier, itemId models.Id) (r_result bool, r_err error) {
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
		itemId,
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
// Duplicates in itemIds have no effect on the result.
// Returns true if all items exist, false otherwise.
func ItemsExist(db DatabaseQuerier, itemIds []models.Id) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if len(itemIds) == 0 {
		return true, nil
	}

	itemIds = algorithms.RemoveDuplicates(itemIds)

	// Set up SQL query
	query := fmt.Sprintf(`
		SELECT
			COUNT(item_id)
		FROM
			items
		WHERE
			item_id IN (%s)
	`, placeholderString(len(itemIds)))

	convertedItemIds := algorithms.Map(itemIds, func(id models.Id) any { return id })
	row := db.QueryRow(query, convertedItemIds...)

	var count int
	err := row.Scan(&count)

	if err != nil {
		return false, err
	}

	return count == len(itemIds), nil
}

// EnsureItemsExist checks if all items with the given IDs exist in the database.
// If any item does not exist, it returns a ErrNoSuchItem.
func EnsureItemsExist(db DatabaseQuerier, itemIds []models.Id) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	itemsExist, err := ItemsExist(db, itemIds)
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
// Duplicates in itemIds are ignored.
// In case of an error, no items are updated.
// This function consists of multiple database interactions, so it must run within a transaction.
func UpdateFreezeStatusOfItems(transaction *TransactionalDatabaseQuerier, itemIds []models.Id, frozen bool) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if len(itemIds) == 0 {
		return nil
	}

	itemIds = algorithms.RemoveDuplicates(itemIds)
	convertedItemIds := algorithms.Map(itemIds, func(id models.Id) any { return id })

	if err := EnsureItemsExist(transaction, itemIds); err != nil {
		return err
	}

	if err := EnsureNoHiddenItems(transaction, itemIds); err != nil {
		return fmt.Errorf("failed to ensure no hidden items: %w", err)
	}

	query := fmt.Sprintf(`
		UPDATE items
		SET frozen = ?
		WHERE item_id IN (%s)
	`, placeholderString(len(itemIds)))

	sqlValues := append([]any{frozen}, convertedItemIds...)

	if _, err := transaction.Exec(query, sqlValues...); err != nil {
		return err
	}

	return nil
}

// UpdateHiddenStatusOfItems updates the hidden status of multiple items at once.
// If any item does not exist, it returns an ErrNoSuchItem.
// If any item is hidden, it returns a ErrItemFrozen error.
// Duplicates in itemIds are ignored.
// In case of an error, no items are updated.
// This function consists of multiple database interactions, so it must run within a transaction.
func UpdateHiddenStatusOfItems(transaction *TransactionalDatabaseQuerier, itemIds []models.Id, hidden bool) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if len(itemIds) == 0 {
		return nil
	}

	itemIds = algorithms.RemoveDuplicates(itemIds)
	convertedItemIds := algorithms.Map(itemIds, func(id models.Id) any { return id })

	if err := EnsureItemsExist(transaction, itemIds); err != nil {
		return err
	}

	// Check if none of the items are frozen
	if err := EnsureNoFrozenItems(transaction, itemIds); err != nil {
		return err
	}

	query := fmt.Sprintf(`
		UPDATE items
		SET hidden = ?
		WHERE item_id IN (%s)
	`, placeholderString(len(itemIds)))

	sqlValues := append([]any{hidden}, convertedItemIds...)

	if _, err := transaction.Exec(query, sqlValues...); err != nil {
		return err
	}

	return nil
}

func partitionItemsBy(database DatabaseQuerier, itemIds []models.Id, columnName string) (*algorithms.Set[models.Id], *algorithms.Set[models.Id], error) {
	query := fmt.Sprintf(`
		SELECT
			item_id,
			%s
		FROM
			items
		WHERE
			item_id IN (%s)
	`, columnName, placeholderString(len(itemIds)))
	convertedItemIds := algorithms.Map(itemIds, func(id models.Id) any { return id })
	rows, err := database.Query(query, convertedItemIds...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer func() { err = errors.Join(err, rows.Close()) }()

	falseSet := algorithms.NewSet[models.Id]()
	trueSet := algorithms.NewSet[models.Id]()
	for rows.Next() {
		var id models.Id
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
func PartitionItemsByHiddenStatus(db DatabaseQuerier, itemIds []models.Id) (r_visible *algorithms.Set[models.Id], r_hidden *algorithms.Set[models.Id], r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	return partitionItemsBy(db, itemIds, "hidden")
}

// PartitionItemsByFrozenStatus partitions the given item IDs into two sets: one for nonfrozen items and one for frozen items.
// If an item ID does not exist in the database, it is ignored.
func PartitionItemsByFrozenStatus(db DatabaseQuerier, itemIds []models.Id) (r_nonfrozen *algorithms.Set[models.Id], r_frozen *algorithms.Set[models.Id], r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	return partitionItemsBy(db, itemIds, "frozen")
}

// ContainsHiddenItems checks if any of the given items are hidden.
// It returns true if at least one item is hidden, and false otherwise.
// It is not an error when nonexistent items are passed in, they are simply ignored.
func ContainsHiddenItems(database DatabaseQuerier, itemIds []models.Id) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if len(itemIds) == 0 {
		return false, nil
	}

	_, hidden, err := PartitionItemsByHiddenStatus(database, itemIds)
	if err != nil {
		return false, err
	}

	containsHidden := hidden.Len() > 0

	return containsHidden, nil
}

// ContainsFrozenItems checks if any of the given items are frozen.
// It returns true if at least one item is frozen, and false otherwise.
// It is not an error when nonexistent items are passed in, they are simply ignored.
func ContainsFrozenItems(database DatabaseQuerier, itemIds []models.Id) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if len(itemIds) == 0 {
		return false, nil
	}

	_, frozen, err := PartitionItemsByFrozenStatus(database, itemIds)
	if err != nil {
		return false, err
	}

	containsFrozen := frozen.Len() > 0

	return containsFrozen, nil
}

// IsItemFrozen checks if none of the items is frozen.
// If one or more items are frozen, it returns an ErrItemFrozen error.
func EnsureNoFrozenItems(qh DatabaseQuerier, itemIds []models.Id) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	containsFrozen, err := ContainsFrozenItems(qh, itemIds)

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
func EnsureNoHiddenItems(qh DatabaseQuerier, itemIds []models.Id) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	containsHidden, err := ContainsHiddenItems(qh, itemIds)

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
func IsItemFrozen(db DatabaseQuerier, itemId models.Id) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	nonFrozen, frozen, err := PartitionItemsByFrozenStatus(db, []models.Id{itemId})
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

	return false, fmt.Errorf("failed to check if item %d is frozen: %w", itemId, dberr.ErrNoSuchItem)
}

// IsItemHidden checks whether the item with the given ID is hidden.
// ErrNoSuchItem is returned if the item does not exist.
func IsItemHidden(db DatabaseQuerier, itemId models.Id) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	unhidden, hidden, err := PartitionItemsByHiddenStatus(db, []models.Id{itemId})
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

	return false, fmt.Errorf("failed to check if item %d is hidden: %w", itemId, dberr.ErrNoSuchItem)
}

// RemoveItemWithId removes the item with the given ID from the database.
// This function breaks the monotonicity invariant upon which other functionality relies,
// so it should be used with caution.
// If the item does not exist, ErrNoSuchItem is returned.
// If the item has been sold, ErrItemSold is returned and the item remains in the database.
// If the item is frozen, ErrItemFrozen is returned and the item remains in the database.
func RemoveItemWithId(db DatabaseQuerier, itemId models.Id) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if err := EnsureItemsExist(db, []models.Id{itemId}); err != nil {
		return fmt.Errorf("failed to remove item with id %d: %w", itemId, dberr.ErrNoSuchItem)
	}

	if err := EnsureNoFrozenItems(db, []models.Id{itemId}); err != nil {
		return fmt.Errorf("failed to remove item with id %d: %w", itemId, dberr.ErrItemFrozen)
	}

	_, err := db.Exec(
		`
			DELETE FROM items
			WHERE item_id = $1
		`,
		itemId,
	)
	if err != nil {
		sold, err2 := HasAnyBeenSold(db, []models.Id{itemId})
		if err2 != nil {
			return err
		}
		if sold {
			return fmt.Errorf("failed to remove item with id %d: %w", itemId, dberr.ErrItemSold)
		}

		return fmt.Errorf("failed to remove item with id %d: %w", itemId, err)
	}

	return nil
}

// ItemUpdate represents the fields that can be updated in an item.
// It is used by the function UpdateItem.
type ItemUpdate struct {
	AddedAt      *models.Timestamp    // If nil, the AddedAt field is not updated.
	Description  *string              // If nil, the Description field is not updated.
	PriceInCents *models.MoneyInCents // If nil, the PriceInCents field is not updated.
	CategoryId   *models.Id           // If nil, the CategoryId field is not updated.
	Donation     *bool                // If nil, the Donation field is not updated.
	Charity      *bool                // If nil, the Charity field is not updated.
}

// UpdateItem updates the item with the given ID in the database.
// If the item does not exist, ErrNoSuchItem is returned.
// If the item is frozen, ErrItemFrozen is returned.
// If the item is hidden, ErrItemHidden is returned.
func UpdateItem(db *TransactionalDatabaseQuerier, itemId models.Id, itemUpdate *ItemUpdate) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	item, err := GetItemWithId(db, itemId)
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

	if itemUpdate.CategoryId != nil {
		categoryExists, err := CategoryWithIdExists(db, *itemUpdate.CategoryId)
		if err != nil {
			return err
		}

		if !categoryExists {
			return fmt.Errorf("failed to update item's category to %d", *itemUpdate.CategoryId)
		}

		sqlUpdates = append(sqlUpdates, "item_category_id = ?")
		sqlValues = append(sqlValues, *itemUpdate.CategoryId)
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

	sqlValues = append(sqlValues, itemId)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE item_id = ?", "items", strings.Join(sqlUpdates, ", "))

	if _, err := db.Exec(query, sqlValues...); err != nil {
		return err
	}

	return nil
}

type AddItemFunction func(addedAt models.Timestamp, description string, priceInCents models.MoneyInCents, itemCategoryId models.Id, sellerId models.Id, donation bool, charity bool, frozen bool, hidden bool)

type AddItemsCallback func(addItem AddItemFunction)

// AddItems allows to add multiple items to the database at once.
func AddItems(db DatabaseQuerier, callback AddItemsCallback) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	valuesString := []string{}
	arguments := []any{}
	tupleString := "(?, ?, ?, ?, ?, ?, ?, ?, ?)"

	add := func(addedAt models.Timestamp, description string, priceInCents models.MoneyInCents, itemCategoryId models.Id, sellerId models.Id, donation bool, charity bool, frozen bool, hidden bool) {
		valuesString = append(valuesString, tupleString)
		arguments = append(arguments, addedAt, description, priceInCents, itemCategoryId, sellerId, donation, charity, frozen, hidden)
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
// Returns ErrNoSuchUser is sellerId does not exist.
// Returns ErrWrongRole if sellerId does not refer to a seller.
func DoesSellerHaveFrozenItems(db *TransactionalDatabaseQuerier, sellerId models.Id) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if err := EnsureUserExistsAndHasRole(db, sellerId, models.NewSellerRoleId()); err != nil {
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

	row := db.QueryRow(query, sellerId)

	var dummy uint64
	if err := row.Scan(&dummy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// MoveItemsToNewSeller moves all items belonging to oldSellerId to newSellerId.
// Does not check for frozen or hidden items!
// Returns ErrNoSuchUser if oldSellerId or newSellerId do not exist.
// Returns ErrWrongRole if oldSellerId or newSellerId do not refer to sellers.
func MoveItemsToNewSeller(db *TransactionalDatabaseQuerier, oldSellerId models.Id, newSellerId models.Id) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	// oldSellerId must refer to seller
	if err := EnsureUserExistsAndHasRole(db, oldSellerId, models.NewSellerRoleId()); err != nil {
		return err
	}

	// newSellerId must refer to seller
	if err := EnsureUserExistsAndHasRole(db, newSellerId, models.NewSellerRoleId()); err != nil {
		return err
	}

	query := `
		UPDATE items
		SET seller_id = ?
		WHERE seller_id = ?
	`
	if _, err := db.Exec(query, newSellerId, oldSellerId); err != nil {
		return err
	}

	return nil
}
