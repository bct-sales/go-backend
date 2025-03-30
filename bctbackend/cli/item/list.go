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

func ListItems(databasePath string) (err error) {
	db, err := database.ConnectToDatabase(databasePath)

	if err != nil {
		return err
	}

	defer func() { err = errors.Join(err, db.Close()) }()

	items := []*models.Item{}
	if err := queries.GetItems(db, queries.CollectTo(&items)); err != nil {
		err = fmt.Errorf("error while listing items: %v", err)
		return err
	}

	itemCount := len(items)

	if itemCount > 0 {
		if err := formatting.PrintItems(items); err != nil {
			err = fmt.Errorf("error while rendering table: %v", err)
			return err
		}

		fmt.Printf("Number of items listed: %d\n", itemCount)

		err = nil
		return err
	} else {
		fmt.Println("No items found")

		err = nil
		return err
	}
}
