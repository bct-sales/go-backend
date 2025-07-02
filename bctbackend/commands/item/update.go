package item

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/MakeNowJust/heredoc"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type updateItemCommand struct {
	common.Command
	itemId       uint64
	description  string
	priceInCents uint64
	categoryId   uint64
	donation     bool
	charity      bool
	noDonation   bool
	noCharity    bool
}

func NewUpdateItemCommand() *cobra.Command {
	var command *updateItemCommand
	command = &updateItemCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "update <item-id>",
				Short: "Updates an item",
				Long: heredoc.Doc(`
				This command updates an existing item in the database.
			   `),
				Args: cobra.NoArgs,
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute(args)
				},
			},
		},
	}

	command.CobraCommand.Flags().Uint64Var(&command.itemId, "id", 0, "ID of the item to update")
	command.CobraCommand.Flags().StringVar(&command.description, "description", "", "New description for the item")
	command.CobraCommand.Flags().Uint64Var(&command.priceInCents, "price", 0, "New price in cents for the item")
	command.CobraCommand.Flags().Uint64Var(&command.categoryId, "category", 0, "New category ID for the item")
	command.CobraCommand.Flags().BoolVar(&command.donation, "donation", false, "Set item as a donation")
	command.CobraCommand.Flags().BoolVar(&command.donation, "no-donation", false, "Unset item as a donation")
	command.CobraCommand.Flags().BoolVar(&command.charity, "charity", false, "Set item as a charity item")
	command.CobraCommand.Flags().BoolVar(&command.charity, "no-charity", false, "Unset item as a charity item")

	command.CobraCommand.MarkFlagRequired("id")
	command.CobraCommand.MarkFlagsMutuallyExclusive("donation", "no-donation")
	command.CobraCommand.MarkFlagsMutuallyExclusive("charity", "no-charity")

	return command.AsCobraCommand()
}

func (c *updateItemCommand) execute(args []string) error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		if err := c.updateItem(db); err != nil {
			return err
		}

		if err := c.showUpdatedItem(db); err != nil {
			return err
		}

		return nil
	})
}

func (c *updateItemCommand) updateItem(db *sql.DB) error {
	var description *string
	var priceInCents *models.MoneyInCents
	var categoryId *models.Id
	var donation *bool
	var charity *bool

	if c.CobraCommand.Flags().Changed("description") {
		description = &c.description
	}

	if c.CobraCommand.Flags().Changed("price") {
		value := models.MoneyInCents(c.priceInCents)
		priceInCents = &value
	}

	if c.CobraCommand.Flags().Changed("category") {
		value := models.Id(c.categoryId)
		categoryId = &value
	}

	if c.CobraCommand.Flags().Changed("donation") {
		donation = &c.donation
	}

	if c.CobraCommand.Flags().Changed("charity") {
		charity = &c.charity
	}

	if c.CobraCommand.Flags().Changed("no-donation") {
		value := false
		donation = &value
	}

	if c.CobraCommand.Flags().Changed("no-charity") {
		value := false
		charity = &value
	}

	itemUpdate := queries.ItemUpdate{
		Description:  description,
		PriceInCents: priceInCents,
		CategoryId:   categoryId,
		Charity:      charity,
		Donation:     donation,
		AddedAt:      nil,
	}

	err := queries.UpdateItem(db, models.Id(c.itemId), &itemUpdate)
	if err != nil {
		c.PrintErrorf("Failed to update item\n")
		return err
	}

	c.Printf("Item updated successfully\n")
	return nil
}

func (c *updateItemCommand) showUpdatedItem(db *sql.DB) error {
	itemId := models.Id(c.itemId)
	categoryNameTable, err := c.GetCategoryNameTable(db)
	if err != nil {
		c.PrintErrorf("Failed to get category name table\n")
		return err
	}

	item, err := queries.GetItemWithId(db, itemId)
	if err != nil {
		c.PrintErrorf("Failed to get item back from database\n")
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
