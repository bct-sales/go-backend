package item

import (
	"bctbackend/cli/formatting"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"
)

func ShowItem(databasePath string, itemId models.Id) (r_err error) {
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	categoryTable, err := queries.GetCategoryNameTable(db)
	if err != nil {
		return fmt.Errorf("failed to get category name table: %w", err)
	}

	err = formatting.PrintItem(db, categoryTable, itemId)
	if err != nil {
		return fmt.Errorf("failed to print item: %w", err)
	}

	return nil
}
