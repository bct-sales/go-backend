package queries

import (
	models "bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"
)

func AddCategory(db *sql.DB, categoryId models.Id, categoryName string) error {
	if !models.IsValidCategoryName(categoryName) {
		return &InvalidCategoryNameError{}
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
				return &CategoryIdAlreadyInUseError{}
			}
		}

		return err
	}

	return nil
}

func CategoryWithIdExists(
	db *sql.DB,
	categoryId models.Id) (bool, error) {

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

		return false, err
	}

	return true, nil
}

func GetCategories(db *sql.DB) (r_result []*models.ItemCategory, r_err error) {
	rows, err := db.Query(
		`
			SELECT item_category_id, name
			FROM item_categories
			ORDER BY item_category_id
		`,
	)
	if err != nil {
		return nil, err
	}

	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	categories := []*models.ItemCategory{}

	for rows.Next() {
		var category models.ItemCategory

		err := rows.Scan(
			&category.CategoryId,
			&category.Name,
		)

		if err != nil {
			return nil, err
		}

		categories = append(categories, &category)
	}

	return categories, nil
}

func GetCategoryNameTable(db *sql.DB) (map[models.Id]string, error) {
	categories, err := GetCategories(db)

	if err != nil {
		return nil, err
	}

	result := make(map[models.Id]string)

	for _, category := range categories {
		result[category.CategoryId] = category.Name
	}

	return result, nil
}

func GetCategoryCounts(db *sql.DB, includeHiddenItems bool) (r_counts map[models.Id]int, r_err error) {
	var hiddenStrategy ItemSelection
	if includeHiddenItems {
		hiddenStrategy = AllItems
	} else {
		hiddenStrategy = OnlyVisibleItems
	}
	itemsTable := ItemsTableFor(hiddenStrategy)

	query := fmt.Sprintf(`
		SELECT item_categories.item_category_id, COUNT(i.item_id)
		FROM item_categories
		LEFT JOIN %s i ON item_categories.item_category_id = i.item_category_id
		GROUP BY item_categories.item_category_id
	`, itemsTable)

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() { r_err = errors.Join(r_err, rows.Close()) }()

	counts := make(map[models.Id]int)

	for rows.Next() {
		var id models.Id
		var count int

		err := rows.Scan(
			&id,
			&count,
		)
		if err != nil {
			return nil, err
		}

		counts[id] = count
	}

	return counts, err
}
