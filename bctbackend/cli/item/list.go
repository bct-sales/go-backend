package item

import (
	"bctbackend/cli/formatting"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

func ListItems(databasePath string, showHidden bool) (r_err error) {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return err
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	var hiddenStrategy queries.ItemSelection
	if showHidden {
		hiddenStrategy = queries.AllItems
	} else {
		hiddenStrategy = queries.OnlyVisibleItems
	}

	items := []*models.Item{}
	if err := queries.GetItems(db, queries.CollectTo(&items), hiddenStrategy); err != nil {
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
