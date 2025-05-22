package csv

import (
	"bctbackend/database"
	dbcsv "bctbackend/database/csv"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"os"

	_ "modernc.org/sqlite"
)

func ExportItems(databasePath string, includeHidden bool) (r_err error) {
	itemSelection := queries.ItemSelectionFromBool(includeHidden)

	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return err
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	items := []*models.Item{}
	if err := queries.GetItems(db, queries.CollectTo(&items), itemSelection); err != nil {
		return err
	}

	categoryTable, err := queries.GetCategoryNameTable(db)
	if err != nil {
		return err
	}

	err = dbcsv.FormatItemsAsCSV(items, categoryTable, os.Stdout)
	if err != nil {
		return err
	}

	return nil
}
