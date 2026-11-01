package queries

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	models "bctbackend/database/models"
	"bctbackend/database/queries/internal/isql"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// AddCategory adds a new category with the given name to the database.
// Returns the ID of the newly created category.
// Returns ErrInvalidCategoryName if the category name is invalid.
// Returns ErrDuplicateCategoryName if there already exists a category with that name.
func AddCategory(ctx context.Context, db DatabaseQuerier, categoryName string) (r_result models.ID, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if !models.IsValidCategoryName(categoryName) {
		return 0, dberr.ErrInvalidCategoryName
	}

	id, err := isql.New(db).AddCategory(ctx, categoryName)
	if err != nil {
		if nameExists, nameErr := CategoryWithNameExists(ctx, db, categoryName); nameErr == nil && nameExists {
			return 0, fmt.Errorf("failed to insert category: %w", dberr.ErrDuplicateCategoryName)
		}

		return 0, fmt.Errorf("failed to insert category: %w", err)
	}

	return id, nil
}

// AddCategoryWithID adds a new category with the given ID and name to the database.
// If the category name is invalid, it returns an ErrInvalidCategoryName error.
func AddCategoryWithID(ctx context.Context, db DatabaseQuerier, categoryID models.ID, categoryName string) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if !models.IsValidCategoryName(categoryName) {
		return dberr.ErrInvalidCategoryName
	}

	err := isql.New(db).AddCategoryWithID(ctx, isql.AddCategoryWithIDParams{
		ItemCategoryID: categoryID,
		Name:           categoryName,
	})
	if err != nil {
		inUse, err := CategoryWithIDExists(ctx, db, categoryID)
		if err == nil && inUse {
			return fmt.Errorf("failed to add category with id %d: %w", categoryID, dberr.ErrIDAlreadyInUse)
		}

		if categoryExists, existenceErr := CategoryWithNameExists(ctx, db, categoryName); existenceErr == nil && categoryExists {
			return fmt.Errorf("failed to insert category: %w", dberr.ErrDuplicateCategoryName)
		}

		return fmt.Errorf("failed to insert category with id %d: %w", categoryID, err)
	}

	return nil
}

// CategoryWithIDExists checks if a category with the given ID exists in the database.
// Returns true if such a category exists, false otherwise.
func CategoryWithIDExists(ctx context.Context, db DatabaseQuerier, categoryID models.ID) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	_, err := isql.New(db).CategoryWithIDExists(ctx, categoryID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("failed to check if category with id %d exists: %w", categoryID, err)
	}

	return true, nil
}

// ListCategories retrieves all categories from the database.
// The categories are returned in ascending order by their ID.
func ListCategories(ctx context.Context, db DatabaseQuerier) (r_result []*models.ItemCategory, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	categories, err := isql.New(db).ListCategories(ctx)
	if err != nil {
		return nil, err
	}

	result := algorithms.Map(categories, func(c isql.ItemCategory) *models.ItemCategory {
		return &models.ItemCategory{
			CategoryID: models.ID(c.ItemCategoryID),
			Name:       c.Name,
		}
	})

	return result, nil
}

// GetCategoryNameTable retrieves the IDs and names of all categories from the database.
// Returns a map where the keys are category IDs and the values are category names.
func GetCategoryNameTable(ctx context.Context, db DatabaseQuerier) (r_result map[models.ID]string, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	categories, err := ListCategories(ctx, db)
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
func CountItemsPerCategory(ctx context.Context, database DatabaseQuerier, itemSelection ItemSelection) (r_counts map[models.ID]int, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	querier := isql.New(database)
	result := make(map[models.ID]int)

	if itemSelection == AllItems {
		rows, err := querier.CountItemsPerCategory(ctx)

		if err != nil {
			return nil, err
		}

		for _, row := range rows {
			result[row.ItemCategoryID] = int(row.Count)
		}
	} else {
		rows, err := querier.CountItemsPerCategoryByHidden(ctx, itemSelection == OnlyHiddenItems)

		if err != nil {
			return nil, err
		}

		for _, row := range rows {
			result[row.ItemCategoryID] = int(row.Count)
		}
	}

	return result, nil
}

// CountSoldItemsPerCategory counts the number of sold items for each category.
// Returns a map where the keys are category IDs and the values are the counts of sold items in that category.
func CountSoldItemsPerCategory(ctx context.Context, database DatabaseQuerier) (r_counts map[models.ID]int, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	rows, err := isql.New(database).CountSoldItemsPerCategory(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get category counts: %w", err)
	}

	counts := make(map[models.ID]int)

	for _, row := range rows {
		counts[row.ItemCategoryID] = int(row.Count)
	}

	return counts, nil
}

// CategoryWithNameExists checks if there exists a category with the given name.
func CategoryWithNameExists(ctx context.Context, db DatabaseQuerier, categoryName string) (r_result bool, r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	_, err := isql.New(db).CategoryWithNameExists(ctx, categoryName)
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
func RenameCategory(ctx context.Context, db DatabaseQuerier, categoryID models.ID, newCategoryName string) (r_err error) {
	defer func() {
		r_err = dberr.WrapError(r_err)
	}()

	if !models.IsValidCategoryName(newCategoryName) {
		return dberr.ErrInvalidCategoryName
	}

	idExists, idExistsErr := CategoryWithIDExists(ctx, db, categoryID)
	if idExistsErr != nil {
		return idExistsErr
	}
	if !idExists {
		return dberr.ErrNoSuchCategory
	}

	nameExists, nameExistsErr := CategoryWithNameExists(ctx, db, newCategoryName)
	if nameExistsErr != nil {
		return nameExistsErr
	}
	if nameExists {
		return dberr.ErrDuplicateCategoryName
	}

	if err := isql.New(db).RenameCategory(ctx, isql.RenameCategoryParams{
		Name:           newCategoryName,
		ItemCategoryID: categoryID,
	}); err != nil {
		return fmt.Errorf("failed to update category name: %w", err)
	}

	return nil
}
