package queries

import (
	models "bctbackend/database/models"
	"database/sql"
)

func CategoryWithIdExists(
	db *sql.DB,
	categoryId models.Id) bool {

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

	return err == nil
}

func GetCategories(db *sql.DB) ([]models.ItemCategory, error) {
	rows, err := db.Query(
		`
			SELECT item_category_id, name
			FROM item_categories
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	categories := []models.ItemCategory{}

	for rows.Next() {
		var category models.ItemCategory

		err := rows.Scan(
			&category.CategoryId,
			&category.Name,
		)

		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	return categories, nil
}

func GetCategoryMap(db *sql.DB) (map[models.Id]string, error) {
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

func GetCategoryCounts(db *sql.DB) ([]models.ItemCategoryCount, error) {
	rows, err := db.Query(
		`
			SELECT item_category_id, item_category_name, count
			FROM item_category_counts
			ORDER BY item_category_id
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	counts := []models.ItemCategoryCount{}

	for rows.Next() {
		var count models.ItemCategoryCount

		err := rows.Scan(
			&count.CategoryId,
			&count.Name,
			&count.Count,
		)

		if err != nil {
			return nil, err
		}

		counts = append(counts, count)
	}

	return counts, nil
}
