package item

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

func FreezeItem(
	databasePath string,
	itemId models.Id) (r_err error) {

	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return err
	}

	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	if err := queries.FreezeItem(db, itemId); err != nil {
		return err
	}

	fmt.Println("Item frozen successfully")
	return nil
}
