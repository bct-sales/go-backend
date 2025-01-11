package item

import (
	"bctbackend/cli/formatting"
	"bctbackend/database"
	"bctbackend/database/models"

	_ "modernc.org/sqlite"
)

func ShowItem(databasePath string, itemId models.Id) error {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return err
	}

	defer db.Close()

	err = formatting.PrintItem(db, itemId)

	if err != nil {
		return err
	}

	return nil
}
