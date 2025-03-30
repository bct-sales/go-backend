package item

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

func RemoveItem(
	databasePath string,
	itemId models.Id) error {

	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return err
	}

	defer func() { err = errors.Join(err, db.Close()) }()

	if err := queries.RemoveItemWithId(db, itemId); err != nil {
		return err
	}

	fmt.Println("Item removed successfully")
	err = nil
	return err
}
