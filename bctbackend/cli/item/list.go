package item

import (
	"bctbackend/cli/formatting"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"fmt"

	_ "modernc.org/sqlite"
)

func ListItems(databasePath string) error {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return err
	}

	defer db.Close()

	items := []*models.Item{}
	if err := queries.GetItems(db, queries.CollectTo(&items)); err != nil {
		return fmt.Errorf("error while listing items: %v", err)
	}

	itemCount := len(items)

	if itemCount > 0 {
		if err := formatting.PrintItems(items); err != nil {
			return fmt.Errorf("error while rendering table: %v", err)
		}

		fmt.Printf("Number of items listed: %d\n", itemCount)

		return nil
	} else {
		fmt.Println("No items found")

		return nil
	}
}
