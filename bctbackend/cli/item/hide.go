package item

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

func HideItem(databasePath string, itemId models.Id) (r_err error) {
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	if err := queries.UpdateHiddenStatusOfItems(db, []models.Id{itemId}, true); err != nil {
		return fmt.Errorf("failed to update database: %w", err)
	}

	fmt.Println("Item hidden successfully")
	return nil
}
