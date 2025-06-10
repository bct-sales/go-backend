package item

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type showItemCommand struct {
	common.Command
}

func NewItemShowCommand() *cobra.Command {
	var command *showItemCommand

	command = &showItemCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "show <item-id>",
				Short: "Show item info",
				Long:  `This command shows detailed information about a specific item.`,
				Args:  cobra.ExactArgs(1), // Expect exactly one argument (the item ID)
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute(args)
				},
			},
		},
	}

	return command.CobraCommand
}

func (c *showItemCommand) execute(args []string) error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		itemId, err := c.ParseItemId(args[0])
		if err != nil {
			c.PrintErrorf("Invalid item ID\n")
			return err
		}

		err = c.printItem(db, itemId)
		if err != nil {
			return err
		}

		return nil
	})
}

func (c *showItemCommand) printItem(db *sql.DB, itemId models.Id) error {
	categoryNameTable, err := c.GetCategoryNameTable(db)
	if err != nil {
		return err
	}

	item, err := queries.GetItemWithId(db, itemId)
	if err != nil {
		return fmt.Errorf("failed to get item with id %d: %w", itemId, err)
	}

	categoryName, ok := categoryNameTable[item.CategoryID]
	if !ok {
		panic("Bug: item has nonexistent category")
	}

	tableData := pterm.TableData{
		{"Property", "Value"},
		{"Description", item.Description},
		{"Price", item.PriceInCents.DecimalNotation()},
		{"Category", categoryName},
		{"Seller", item.SellerID.String()},
		{"Donation", strconv.FormatBool(item.Donation)},
		{"Charity", strconv.FormatBool(item.Charity)},
		{"Added At", item.AddedAt.FormattedDateTime()},
	}

	err = pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render()
	if err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}
