package item

import (
	"bctbackend/cli/formatting"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"

	"log/slog"

	_ "modernc.org/sqlite"
)

func AddItem(
	databasePath string,
	description string,
	priceInCents models.MoneyInCents,
	categoryId models.Id,
	sellerId models.Id,
	donation bool,
	charity bool) (r_err error) {

	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return err
	}

	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	timestamp := models.Now()

	// We do not want to check the validity for frozen/hidden here, so we just set them to false
	addedItemId, err := queries.AddItem(db, timestamp, description, priceInCents, categoryId, sellerId, donation, charity, false, false)
	if err != nil {
		return err
	}
	fmt.Println("Item added successfully")

	categoryTable, err := queries.GetCategoryNameTable(db)
	if err != nil {
		slog.Error("An error occurred while trying to get the category map; item is still added to the database", "error", err)
		return nil
	}

	err = formatting.PrintItem(db, categoryTable, addedItemId)
	if err != nil {
		slog.Error("An error occurred while trying to format the output; item is still added to the database", "added item id", addedItemId, "error", err)
		return nil
	}

	return nil
}
