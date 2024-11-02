package item

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"

	_ "modernc.org/sqlite"
)

func RemoveItem(
	databasePath string,
	itemId models.Id) error {

	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return err
	}

	defer db.Close()

	if err := queries.RemoveItemWithId(db, itemId); err != nil {
		return err
	}

	return nil
}
