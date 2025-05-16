package queries

import (
	models "bctbackend/database/models"
	"database/sql"
	"errors"
	"fmt"
)

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

func GetCategoryCounts(db *sql.DB, includeHiddenItems bool) (counts []models.ItemCategoryCount, err error) {
	var whereClause string
	if includeHiddenItems {
		whereClause = ""
	} else {
		whereClause = "WHERE hidden = false"
	}

	query := fmt.Sprintf(`
		SELECT
			item_categories.item_category_id as item_category_id,
			item_categories.name as item_category_name,
			COUNT(i.item_id) AS count
		FROM item_categories
		LEFT JOIN (SELECT * FROM items %s) as i ON item_categories.item_category_id = i.item_category_id
		GROUP BY item_categories.item_category_id
	`, whereClause)

	rows, err := db.Query(query)

	if err != nil {
		return nil, err
	}

	defer func() { err = errors.Join(err, rows.Close()) }()

	counts = []models.ItemCategoryCount{}

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

	err = nil
	return counts, err
}
