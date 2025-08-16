package queries

import (
	dberr "bctbackend/database/errors"
	models "bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"
)

// AddCategory adds a new category with the given name to the database.
// Returns the ID of the newly created category.
// Returns ErrInvalidCategoryName if the category name is invalid.
// Returns ErrDuplicateCategoryName if there already exists a category with that name.
func AddCategory(db DatabaseQuerier, categoryName string) (r_result models.Id, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if !models.IsValidCategoryName(categoryName) {
		return 0, dberr.ErrInvalidCategoryName
	}

	query := `
		INSERT INTO item_categories (name)
		VALUES ($1)
		RETURNING item_category_id
	`
	result, err := db.Exec(query, categoryName)
	if err != nil {
		if categoryExists, existenceErr := CategoryWithNameExists(db, categoryName); existenceErr == nil && categoryExists {
			return 0, fmt.Errorf("failed to insert category: %w", dberr.ErrDuplicateCategoryName)
		}

		return 0, fmt.Errorf("failed to insert category: %w", err)
	}

	categoryId, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to determine id of inserted category: %w", err)
	}

	return models.Id(categoryId), nil
}

// AddCategoryWithId adds a new category with the given ID and name to the database.
// If the category name is invalid, it returns an ErrInvalidCategoryName error.
func AddCategoryWithId(db DatabaseQuerier, categoryId models.Id, categoryName string) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if !models.IsValidCategoryName(categoryName) {
		return dberr.ErrInvalidCategoryName
	}

	_, err := db.Exec(
		`
			INSERT INTO item_categories (item_category_id, name)
			VALUES ($1, $2)
			RETURNING item_category_id
		`,
		categoryId,
		categoryName,
	)
	if err != nil {
		{
			inUse, err := CategoryWithIdExists(db, categoryId)
			if err == nil && inUse {
				return fmt.Errorf("failed to add category with id %d: %w", categoryId, dberr.ErrIdAlreadyInUse)
			}
		}

		if categoryExists, existenceErr := CategoryWithNameExists(db, categoryName); existenceErr == nil && categoryExists {
			return fmt.Errorf("failed to insert category: %w", dberr.ErrDuplicateCategoryName)
		}

		return fmt.Errorf("failed to insert category with id %d: %w", categoryId, err)
	}

	return nil
}

// CategoryWithIdExists checks if a category with the given ID exists in the database.
// Returns true if such a category exists, false otherwise.
func CategoryWithIdExists(db DatabaseQuerier, categoryId models.Id) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	row := db.QueryRow(
		`
			SELECT 1
			FROM item_categories
			WHERE item_category_id = $1
		`,
		categoryId,
	)

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

	rows, err := db.Query(
		`
			SELECT item_category_id, name
			FROM item_categories
			ORDER BY item_category_id
		`,
	)
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
func GetCategoryNameTable(db DatabaseQuerier) (r_result map[models.Id]string, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	categories, err := GetCategories(db)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	result := make(map[models.Id]string)

	for _, category := range categories {
		result[category.CategoryID] = category.Name
	}

	return result, nil
}

// CountItemsPerCategory retrieves the count of items in each category.
// Returns a map where the keys are category IDs and the values are the counts of items in that category.
// The itemSelection parameter allows filtering items based on specific criteria.
func CountItemsPerCategory(database DatabaseQuerier, itemSelection ItemSelection) (r_counts map[models.Id]int, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	itemsTable := ItemsTableFor(itemSelection)

	query := fmt.Sprintf(`
		SELECT item_categories.item_category_id, COUNT(i.item_id)
		FROM item_categories
		LEFT JOIN %s i ON item_categories.item_category_id = i.item_category_id
		GROUP BY item_categories.item_category_id
	`, itemsTable)

	rows, err := database.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get category counts: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	counts := make(map[models.Id]int)

	for rows.Next() {
		var categoryId models.Id
		var count int

		err := rows.Scan(
			&categoryId,
			&count,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		counts[categoryId] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred while iterating over rows: %w", err)
	}

	return counts, nil
}

func CategoryWithNameExists(db DatabaseQuerier, categoryName string) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	row := db.QueryRow(
		`
			SELECT 1
			FROM item_categories
			WHERE name = $1
		`,
		categoryName,
	)

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
