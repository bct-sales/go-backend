package user

import (
	"bctbackend/database"
	"bctbackend/database/queries"
	"errors"
	"fmt"
	"strconv"

	"github.com/pterm/pterm"
	_ "modernc.org/sqlite"
)

func ListCategoryCounts(databasePath string) (r_err error) {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return err
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	categoryCounts, err := queries.GetCategoryCounts(db, false)
	if err != nil {
		return err
	}

	categoryTable, err := queries.GetCategoryNameTable(db)
	if err != nil {
		return err
	}

	tableData := pterm.TableData{
		{"ID", "Name", "Count"},
	}

	for categoryId, categoryCount := range categoryCounts {
		categoryNameString, ok := categoryTable[categoryId]
		if !ok {
			return fmt.Errorf("category ID %d not found in category table", categoryId)
		}
		categoryIdString := strconv.FormatInt(categoryId, 10)
		count := strconv.Itoa(categoryCount)

		tableData = append(tableData, []string{
			categoryIdString,
			categoryNameString,
			count,
		})
	}

	if err = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render(); err != nil {
		return fmt.Errorf("error while rendering table: %w", err)
	}

	return nil
}
