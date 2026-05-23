package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/meta"
	models "bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
)

// AddCategory adds a new category with the given name to the database.
// Returns the ID of the newly created category.
// Returns ErrInvalidCategoryName if the category name is invalid.
// Returns ErrDuplicateCategoryName if there already exists a category with that name.
func AddCategory(db DatabaseQuerier, categoryName string) (r_result models.ID, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if !models.IsValidCategoryName(categoryName) {
		return 0, dberr.ErrInvalidCategoryName
	}

	query := squirrel.Insert(meta.ItemCategory.Table).Columns(meta.ItemCategory.Name).Values(categoryName).Suffix("RETURNING " + meta.ItemCategory.ItemCategoryID)
	result, err := query.RunWith(db).Exec()
	if err != nil {
		if categoryExists, existenceErr := CategoryWithNameExists(db, categoryName); existenceErr == nil && categoryExists {
			return 0, fmt.Errorf("failed to insert category: %w", dberr.ErrDuplicateCategoryName)
		}

		return 0, fmt.Errorf("failed to insert category: %w", err)
	}

	categoryID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to determine id of inserted category: %w", err)
	}

	return models.ID(categoryID), nil
}

// AddCategoryWithID adds a new category with the given ID and name to the database.
// If the category name is invalid, it returns an ErrInvalidCategoryName error.
func AddCategoryWithID(db DatabaseQuerier, categoryID models.ID, categoryName string) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if !models.IsValidCategoryName(categoryName) {
		return dberr.ErrInvalidCategoryName
	}

	query := squirrel.Insert(meta.ItemCategory.Table).Columns(meta.ItemCategory.ItemCategoryID, meta.ItemCategory.Name).Values(categoryID, categoryName)
	if _, err := query.RunWith(db).Exec(); err != nil {
		inUse, err := CategoryWithIDExists(db, categoryID)
		if err == nil && inUse {
			return fmt.Errorf("failed to add category with id %d: %w", categoryID, dberr.ErrIDAlreadyInUse)
		}

		if categoryExists, existenceErr := CategoryWithNameExists(db, categoryName); existenceErr == nil && categoryExists {
			return fmt.Errorf("failed to insert category: %w", dberr.ErrDuplicateCategoryName)
		}

		return fmt.Errorf("failed to insert category with id %d: %w", categoryID, err)
	}

	return nil
}

// CategoryWithIDExists checks if a category with the given ID exists in the database.
// Returns true if such a category exists, false otherwise.
func CategoryWithIDExists(db DatabaseQuerier, categoryID models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	query := squirrel.Select("1").From(meta.ItemCategory.Table).Where(squirrel.Eq{meta.ItemCategory.ItemCategoryID: categoryID})
	row := query.RunWith(db).QueryRow()

	var dummy int
	err := row.Scan(&dummy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("failed to read row: %w", err)
	}

	return true, nil
}

// GetCategories retrieves all categories from the database.
// The categories are returned in ascending order by their ID.
func GetCategories(db DatabaseQuerier) (r_result []*models.ItemCategory, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	query := squirrel.Select(meta.ItemCategory.ItemCategoryID, meta.ItemCategory.Name).From(meta.ItemCategory.Table).OrderBy(meta.ItemCategory.ItemCategoryID)
	rows, err := query.RunWith(db).Query()
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	categories := []*models.ItemCategory{}

	for rows.Next() {
		var category models.ItemCategory

		err := rows.Scan(
			&category.CategoryID,
			&category.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		categories = append(categories, &category)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return categories, nil
}

// GetCategoryNameTable retrieves the IDs and names of all categories from the database.
// Returns a map where the keys are category IDs and the values are category names.
func GetCategoryNameTable(db DatabaseQuerier) (r_result map[models.ID]string, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	categories, err := GetCategories(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	result := make(map[models.ID]string)

	for _, category := range categories {
		result[category.CategoryID] = category.Name
	}

	return result, nil
}

// CountItemsPerCategory retrieves the count of items in each category.
// Returns a map where the keys are category IDs and the values are the counts of items in that category.
// The itemSelection parameter allows filtering items based on specific criteria.
func CountItemsPerCategory(database DatabaseQuerier, itemSelection ItemSelection) (r_counts map[models.ID]int, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	itemsTable := ItemsTableFor(itemSelection)

	query := fmt.Sprintf(`
		SELECT
			item_categories.item_category_id, COUNT(i.item_id)
		FROM
			item_categories
		LEFT JOIN
			%s i ON item_categories.item_category_id = i.item_category_id
		GROUP BY
			item_categories.item_category_id
	`, itemsTable)

	rows, err := database.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get category counts: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	counts := make(map[models.ID]int)

	for rows.Next() {
		var categoryID models.ID
		var count int

		err := rows.Scan(
			&categoryID,
			&count,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		counts[categoryID] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return counts, nil
}

// CountSoldItemsPerCategory counts the number of sold items for each category.
// Returns a map where the keys are category IDs and the values are the counts of sold items in that category.
func CountSoldItemsPerCategory(database DatabaseQuerier) (r_counts map[models.ID]int, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	query := `
		SELECT
			item_categories.item_category_id, COUNT(sold_items.item_id)
		FROM
			item_categories
		LEFT JOIN
			(
					sale_items
				INNER JOIN
					items
				ON
					sale_items.item_id = items.item_id
			) as sold_items
			ON
				item_categories.item_category_id = sold_items.item_category_id
		GROUP BY
			item_categories.item_category_id
	`

	rows, err := database.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get category counts: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	counts := make(map[models.ID]int)

	for rows.Next() {
		var categoryID models.ID
		var count int

		err := rows.Scan(
			&categoryID,
			&count,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		counts[categoryID] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return counts, nil
}

// CategoryWithNameExists checks if there exists a category with the given name.
func CategoryWithNameExists(db DatabaseQuerier, categoryName string) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	query := squirrel.Select("1").From(meta.ItemCategory.Table).Where(squirrel.Eq{meta.ItemCategory.Name: categoryName})
	row := query.RunWith(db).QueryRow()

	var dummy int
	err := row.Scan(&dummy)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("failed to read row: %w", err)
	}

	return true, nil
}

// RenameCategory updates a category's name.
// If the new name is invalid, an ErrInvalidCategoryName is returned.
// If the id is invalid, an ErrNoSuchCategory is returned.
// If the new name is in use by another category, an ErrDuplicateCategoryName is returned.
func RenameCategory(db DatabaseQuerier, categoryID models.ID, newCategoryName string) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if !models.IsValidCategoryName(newCategoryName) {
		return dberr.ErrInvalidCategoryName
	}

	idExists, idExistsErr := CategoryWithIDExists(db, categoryID)
	if idExistsErr != nil {
		return idExistsErr
	}
	if !idExists {
		return dberr.ErrNoSuchCategory
	}

	nameExists, nameExistsErr := CategoryWithNameExists(db, newCategoryName)
	if nameExistsErr != nil {
		return nameExistsErr
	}
	if nameExists {
		return dberr.ErrDuplicateCategoryName
	}

	query := squirrel.Update(meta.ItemCategory.Table).Set(meta.ItemCategory.Name, newCategoryName).Where(squirrel.Eq{meta.ItemCategory.ItemCategoryID: categoryID})
	if _, err := query.RunWith(db).Exec(); err != nil {
		return fmt.Errorf("failed to update category name: %w", err)
	}

	return nil
}
