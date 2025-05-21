package item

import (
	"bctbackend/cli/formatting"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"

	_ "modernc.org/sqlite"
)

func ShowItem(databasePath string, itemId models.Id) (r_err error) {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return err
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	categoryTable, err := queries.GetCategoryNameTable(db)
	if err != nil {
		return err
	}

	err = formatting.PrintItem(db, categoryTable, itemId)
	if err != nil {
		return err
	}

	return nil
}
