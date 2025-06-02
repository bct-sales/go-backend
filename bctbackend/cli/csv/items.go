package csv

import (
	"bctbackend/database"
	dbcsv "bctbackend/database/csv"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func ExportItems(databasePath string, includeHidden bool) (r_err error) {
	itemSelection := queries.ItemSelectionFromBool(includeHidden)

	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	items := []*models.Item{}
	if err := queries.GetItems(db, queries.CollectTo(&items), itemSelection); err != nil {
		return fmt.Errorf("failed to get items: %w", err)
	}

	categoryTable, err := queries.GetCategoryNameTable(db)
	if err != nil {
		return fmt.Errorf("failed to get category name table: %w", err)
	}

	err = dbcsv.FormatItemsAsCSV(items, categoryTable, os.Stdout)
	if err != nil {
		return fmt.Errorf("failed to format items as a CSV: %w", err)
	}

	return nil
}
