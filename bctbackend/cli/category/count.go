package user

import (
	"bctbackend/database"
	"bctbackend/database/queries"
	"errors"
	"fmt"

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

	tableData := pterm.TableData{
		{"ID", "Name", "Count"},
	}

	for _, categoryCount := range categoryCounts {
		categoryIdString := fmt.Sprintf("%d", categoryCount.CategoryId)
		categoryNameString := categoryCount.Name
		count := fmt.Sprintf("%d", categoryCount.Count)

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
