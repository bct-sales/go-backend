package item

import (
	"bctbackend/cli/formatting"
	"bctbackend/commands/common"
	dbcsv "bctbackend/database/csv"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewItemListCommand() *cobra.Command {
	var showHidden bool
	var format string

	command := cobra.Command{
		Use:   "list",
		Short: "List all items",
		Long:  `This command lists all items in the database.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch format {
			case "table":
				return listItemsInTableFormat(cmd, showHidden)
			case "csv":
				return listItemsInCSVFormat(cmd, showHidden)
			default:
				fmt.Fprintf(cmd.ErrOrStderr(), "Invalid format: %s\n", format)
				return fmt.Errorf("unknown format: %s", format)
			}
		},
	}

	command.Flags().BoolVar(&showHidden, "show-hidden", false, "Show hidden items")
	command.Flags().StringVar(&format, "format", "table", "Output format (format, csv)")

	return &command
}

func listItemsInTableFormat(cmd *cobra.Command, showHidden bool) error {
	return common.WithOpenedDatabase(cmd.ErrOrStderr(), func(db *sql.DB) error {
		var hiddenStrategy queries.ItemSelection
		if showHidden {
			hiddenStrategy = queries.AllItems
		} else {
			hiddenStrategy = queries.OnlyVisibleItems
		}

		items := []*models.Item{}
		if err := queries.GetItems(db, queries.CollectTo(&items), hiddenStrategy); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error while getting items: %v\n", err)
			return err
		}

		itemCount := len(items)

		if itemCount > 0 {
			categoryNameTable, err := queries.GetCategoryNameTable(db)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "An error occurred while trying to get the category table: %v\n", err)
				return err
			}

			if err := formatting.PrintItems(categoryNameTable, items); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error while printing items: %v\n", err)
				return err
			}

			fmt.Printf("Number of items listed: %d\n", itemCount)

			return nil
		} else {
			fmt.Println("No items found")

			return nil
		}
	})
}

func listItemsInCSVFormat(cmd *cobra.Command, showHidden bool) error {
	return common.WithOpenedDatabase(cmd.ErrOrStderr(), func(db *sql.DB) error {
		itemSelection := queries.ItemSelectionFromBool(showHidden)

		items := []*models.Item{}
		if err := queries.GetItems(db, queries.CollectTo(&items), itemSelection); err != nil {
			return fmt.Errorf("failed to get items: %w", err)
		}

		categoryNameTable, err := queries.GetCategoryNameTable(db)
		if err != nil {
			return fmt.Errorf("failed to get category name table: %w", err)
		}

		err = dbcsv.FormatItemsAsCSV(items, categoryNameTable, os.Stdout)
		if err != nil {
			return fmt.Errorf("failed to format items as a CSV: %w", err)
		}

		return nil
	})
}
