package item

import (
	"bctbackend/cli/formatting"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"fmt"
	"time"

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
	frozen bool) error {

	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return err
	}

	defer db.Close()

	timestamp := time.Now().Unix()

	addedItemId, err := queries.AddItem(db, timestamp, description, priceInCents, categoryId, sellerId, donation, charity, frozen)

	if err != nil {
		return err
	}

	fmt.Println("Item added successfully")

	err = formatting.PrintItem(db, addedItemId)

	if err != nil {
		fmt.Printf("An error occurred while trying to format the output: %v\n", err)
		fmt.Printf("Item is still added to the database.\n")
	}

	return nil
}
