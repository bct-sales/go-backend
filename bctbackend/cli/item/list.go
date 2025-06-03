package item

import (
	"bctbackend/cli/formatting"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"
)

func ListItems(databasePath string, showHidden bool) (r_err error) {
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
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
		return fmt.Errorf("error while listing items: %w", err)
	}

	itemCount := len(items)

	if itemCount > 0 {
		categoryTable, err := queries.GetCategoryNameTable(db)
		if categoryTable == nil {
			return fmt.Errorf("error while getting category map: %w", err)
		}

		if err := formatting.PrintItems(categoryTable, items); err != nil {
			return fmt.Errorf("error while rendering table: %w", err)
		}

		fmt.Printf("Number of items listed: %d\n", itemCount)

		return nil
	} else {
		fmt.Println("No items found")

		return nil
	}
}
