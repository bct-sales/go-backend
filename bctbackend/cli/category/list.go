package user

import (
	"bctbackend/database"
	"bctbackend/database/queries"
	"errors"
	"fmt"

	"github.com/pterm/pterm"
)

func ListCategories(databasePath string) (r_err error) {
	database, err := database.OpenDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	defer func() { r_err = errors.Join(r_err, database.Close()) }()

	categories, err := queries.GetCategories(database)
	if err != nil {
		return fmt.Errorf("failed to list categories: %w", err)
	}

	tableData := pterm.TableData{
		{"ID", "Name"},
	}

	for _, category := range categories {
		categoryIdString := fmt.Sprintf("%d", category.CategoryID)
		categoryNameString := category.Name

		tableData = append(tableData, []string{
			categoryIdString,
			categoryNameString,
		})
	}

	err = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
	if err != nil {
		return fmt.Errorf("error while rendering table: %w", err)
	}

	return nil
}
