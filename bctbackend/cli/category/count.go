package user

import (
	"bctbackend/database"
	"bctbackend/database/queries"
	"errors"
	"fmt"
	"strconv"

	"github.com/pterm/pterm"
	"github.com/urfave/cli/v2"
)

func ListCategoryCounts(databasePath string, itemSelection queries.ItemSelection) (r_err error) {
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	categoryCounts, err := queries.GetCategoryCounts(db, itemSelection)
	if err != nil {
		return fmt.Errorf("failed to get category counts: %w", err)
	}

	categoryNameTable, err := queries.GetCategoryNameTable(db)
	if err != nil {
		return fmt.Errorf("failed to get category name table: %w", err)
	}

	tableData := pterm.TableData{
		{"ID", "Name", "Count"},
	}

	for categoryId, categoryCount := range categoryCounts {
		categoryNameString, ok := categoryNameTable[categoryId]
		if !ok {
			return cli.Exit(fmt.Sprintf("Bug: unknown category %d", categoryId), 1)
		}
		categoryIdString := categoryId.String()
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
