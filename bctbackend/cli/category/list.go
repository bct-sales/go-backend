package user

import (
	"bctbackend/database/queries"
	"fmt"

	"database/sql"

	"github.com/pterm/pterm"
	_ "modernc.org/sqlite"
)

func ListCategories(databasePath string) error {
	db, err := sql.Open("sqlite", databasePath)

	if err != nil {
		return fmt.Errorf("error while opening database: %v", err)
	}

	defer db.Close()

	categories, err := queries.GetCategories(db)

	if err != nil {
		return fmt.Errorf("error while listing categories: %v", err)
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
		return fmt.Errorf("error while rendering table: %v", err)
	}

	return nil
}
