package item

import (
	"bctbackend/cli/formatting"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"
	"time"

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
	charity bool,
	frozen bool) (err error) {

	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return err
	}

	defer func() { err = errors.Join(err, db.Close()) }()

	timestamp := time.Now().Unix()

	addedItemId, err := queries.AddItem(db, timestamp, description, priceInCents, categoryId, sellerId, donation, charity, frozen)
	if err != nil {
		return err
	}

	fmt.Println("Item added successfully")

	err = formatting.PrintItem(db, addedItemId)

	if err != nil {
		slog.Error("An error occurred while trying to format the output; item is still added to the database", "added item id", addedItemId, "error", err)
		return nil
	}

	return nil
}
