package item

import (
	"bctbackend/commands/common"
	dbcsv "bctbackend/database/csv"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type listItemsCommand struct {
	common.Command
	showHidden bool
	format     string
}

func NewListItemsCommand() *cobra.Command {
	var command *listItemsCommand

	command = &listItemsCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "list",
				Short: "List all items",
				Long:  `This command lists all items in the database.`,
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
			},
		},
	}

	command.CobraCommand.Flags().BoolVar(&command.showHidden, "show-hidden", false, "Show hidden items")
	command.CobraCommand.Flags().StringVar(&command.format, "format", "table", "Output format (format, csv)")

	return command.AsCobraCommand()
}

func (c *listItemsCommand) execute() error {
	switch c.format {
	case "table":
		return c.listItemsInTableFormat()
	case "csv":
		return c.listItemsInCSVFormat()
	default:
		c.PrintErrorf("Invalid format: %s\n", c.format)
		return fmt.Errorf("unknown format: %s", c.format)
	}
}

func (c *listItemsCommand) listItemsInTableFormat() error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		var itemSelection queries.ItemSelection
		if c.showHidden {
			itemSelection = queries.AllItems
		} else {
			itemSelection = queries.OnlyVisibleItems
		}

		items := []*models.Item{}
		if err := queries.GetItems(db, queries.CollectTo(&items), itemSelection, queries.AllRows()); err != nil {
			c.PrintErrorf("Error while getting items: %v\n", err)
			return err
		}

		itemCount := len(items)

		if itemCount > 0 {
			categoryNameTable, err := c.GetCategoryNameTable(db)
			if err != nil {
				return err
			}

			tableData := pterm.TableData{
				{"ID", "Description", "Price", "Category", "Seller", "Donation", "Charity", "Added At", "Frozen", "Hidden"},
			}

			for _, item := range items {
				categoryName, ok := categoryNameTable[item.CategoryID]
				if !ok {
					return dberr.ErrNoSuchCategory
				}

				tableData = append(tableData, []string{
					item.ItemID.String(),
					item.Description,
					item.PriceInCents.DecimalNotation(),
					categoryName,
					item.SellerID.String(),
					strconv.FormatBool(item.Donation),
					strconv.FormatBool(item.Charity),
					item.AddedAt.FormattedDateTime(),
					strconv.FormatBool(item.Frozen),
					strconv.FormatBool(item.Hidden),
				})
			}

			if err := pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render(); err != nil {
				c.PrintErrorf("Error while rendering table\n")
				return fmt.Errorf("failed to render table: %w", err)
			}

			c.Printf("Number of items listed: %d\n", itemCount)
			return nil
		} else {
			c.Printf("No items found")
			return nil
		}
	})
}

func (c *listItemsCommand) listItemsInCSVFormat() error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		itemSelection := queries.ItemSelectionFromBool(c.showHidden)

		items := []*models.Item{}
		if err := queries.GetItems(db, queries.CollectTo(&items), itemSelection, queries.AllRows()); err != nil {
			return fmt.Errorf("failed to get items: %w", err)
		}

		categoryNameTable, err := c.GetCategoryNameTable(db)
		if err != nil {
			return err
		}

		err = dbcsv.FormatItemsAsCSV(items, categoryNameTable, os.Stdout)
		if err != nil {
			c.PrintErrorf("Error while formatting items as CSV\n")
			return fmt.Errorf("failed to format items as a CSV: %w", err)
		}

		return nil
	})
}
