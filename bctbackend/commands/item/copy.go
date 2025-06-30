package item

import (
	"bctbackend/commands/common"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/MakeNowJust/heredoc"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type copyItemCommand struct {
	common.Command
}

func NewCopyItemCommand() *cobra.Command {
	var command *copyItemCommand

	command = &copyItemCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "copy <item-id>",
				Short: "Copies an item",
				Long: heredoc.Doc(`
					This command makes a copy of an existing item in the database.
					The new item will have the same description, price, category, and seller as the original item.
					The added time will be set to the current time.
					The new item will always be unfrozen and visible.
				`),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute(args)
				},
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *copyItemCommand) execute(args []string) error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		itemId, err := models.ParseId(args[0])
		if err != nil {
			c.PrintErrorf("Invalid item ID\n")
			return err
		}

		item, err := queries.GetItemWithId(db, itemId)
		if err != nil {
			c.PrintErrorf("Failed to retrieve item with given ID\n")
			return fmt.Errorf("failed to get item with ID %d: %w", itemId, err)
		}

		timestamp := models.Now()

		copyId, err := queries.AddItem(
			db,
			timestamp,
			item.Description,
			item.PriceInCents,
			item.CategoryID,
			item.SellerID,
			item.Donation,
			item.Charity,
			false,
			false)
		if err != nil {
			c.PrintErrorf("Failed to copy item: %v\n", err)
			return fmt.Errorf("failed to insert copy in database: %w", err)
		}

		if err := c.printItem(db, copyId); err != nil {
			return fmt.Errorf("failed to print copied item: %w", err)
		}

		return nil
	})
}

func (c *copyItemCommand) printItem(db *sql.DB, itemId models.Id) error {
	item, err := queries.GetItemWithId(db, itemId)
	if err != nil {
		c.PrintErrorf("Failed to get item back from database\n")
		return fmt.Errorf("failed to get item with id %d: %w", itemId, err)
	}

	categoryNameTable, err := c.GetCategoryNameTable(db)
	if err != nil {
		return err
	}

	categoryName, ok := categoryNameTable[item.CategoryID]
	if !ok {
		c.PrintErrorf("Bug: category does not have a name in the category name table\n")
		return dberr.ErrNoSuchCategory
	}

	tableData := pterm.TableData{
		{"Property", "Value"},
		{"Item ID", item.ItemID.String()},
		{"Description", item.Description},
		{"Price", item.PriceInCents.DecimalNotation()},
		{"Category", categoryName},
		{"Seller", item.SellerID.String()},
		{"Donation", strconv.FormatBool(item.Donation)},
		{"Charity", strconv.FormatBool(item.Charity)},
		{"Added At", item.AddedAt.FormattedDateTime()},
		{"Frozen", strconv.FormatBool(item.Frozen)},
		{"Hidden", strconv.FormatBool(item.Hidden)},
	}

	err = pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render()
	if err != nil {
		c.PrintErrorf("Failed to render item table\n")
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}
