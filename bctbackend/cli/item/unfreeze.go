package item

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

func UnfreezeItem(
	databasePath string,
	itemId models.Id) (r_err error) {

	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return err
	}

	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	if err := queries.UpdateFreezeStatusOfItems(db, []models.Id{itemId}, false); err != nil {
		return err
	}

	fmt.Println("Item unfrozen successfully")
	return nil
}
