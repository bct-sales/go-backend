package user

import (
	"bctbackend/database"
	"bctbackend/database/queries"
	"errors"
	"fmt"

	"github.com/pterm/pterm"
	_ "modernc.org/sqlite"
)

func ListCategories(databasePath string) error {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return err
	}

	defer func() { err = errors.Join(err, db.Close()) }()

	categories, err := queries.GetCategories(db)
	if err != nil {
		return fmt.Errorf("error while listing categories: %w", err)
	}

	tableData := pterm.TableData{
		{"ID", "Name"},
	}

	for _, category := range categories {
		categoryIdString := fmt.Sprintf("%d", category.CategoryId)
		categoryNameString := category.Name

		tableData = append(tableData, []string{
			categoryIdString,
			categoryNameString,
		})
	}

	err = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	if err != nil {
		err = fmt.Errorf("error while rendering table: %w", err)
		return err
	}

	err = nil
	return err
}
