package item

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"
)

func FreezeItem(databasePath string, itemId models.Id) (r_err error) {
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	if err := queries.UpdateFreezeStatusOfItems(db, []models.Id{itemId}, true); err != nil {
		return fmt.Errorf("failed to update database: %w", err)
	}

	fmt.Println("Item frozen successfully")
	return nil
}
