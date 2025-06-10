package item

import (
	"bctbackend/commands/common"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/MakeNowJust/heredoc"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type addItemCommand struct {
	common.Command
	description  string
	priceInCents int
	categoryId   int
	sellerId     int
	donation     bool
	charity      bool
}

func NewItemAddCommand() *cobra.Command {
	var command *addItemCommand

	command = &addItemCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "add",
				Short: "Add an item",
				Long: heredoc.Doc(`
					This command adds a new item to the database.
					The item will always be unfrozen and visible.
					Freezing/hiding needs to be done separately.
				`),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
			},
		},
	}

	command.CobraCommand.Flags().StringVar(&command.description, "description", "", "Description of the item")
	command.CobraCommand.Flags().IntVar(&command.priceInCents, "price", 0, "Price of the item in cents")
	command.CobraCommand.Flags().IntVar(&command.categoryId, "category", 0, "ID of the category the item belongs to")
	command.CobraCommand.Flags().IntVar(&command.sellerId, "seller", 0, "ID of the seller of the item")
	command.CobraCommand.Flags().BoolVar(&command.donation, "donation", false, "Whether the item is a donation")
	command.CobraCommand.Flags().BoolVar(&command.charity, "charity", false, "Whether the item is for charity")

	if err := command.CobraCommand.MarkFlagRequired("description"); err != nil {
		panic(fmt.Sprintf("failed to mark description flag as required: %v", err))
	}
	if err := command.CobraCommand.MarkFlagRequired("price"); err != nil {
		panic(fmt.Sprintf("failed to mark price flag as required: %v", err))
	}
	if err := command.CobraCommand.MarkFlagRequired("category"); err != nil {
		panic(fmt.Sprintf("failed to mark category flag as required: %v", err))
	}
	if err := command.CobraCommand.MarkFlagRequired("seller"); err != nil {
		panic(fmt.Sprintf("failed to mark seller flag as required: %v", err))
	}

	return command.AsCobraCommand()
}

func (c *addItemCommand) execute() error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		timestamp := models.Now()

		addedItemId, err := queries.AddItem(
			db,
			timestamp,
			c.description,
			models.MoneyInCents(c.priceInCents),
			models.Id(c.categoryId),
			models.Id(c.sellerId),
			c.donation,
			c.charity,
			false,
			false)

		if err != nil {
			if errors.Is(err, dberr.ErrNoSuchCategory) {
				c.PrintErrorf("No such category with ID %d\n", c.categoryId)
				return err
			} else if errors.Is(err, dberr.ErrNoSuchUser) {
				c.PrintErrorf("No user with ID %d\n", c.sellerId)
				return err
			} else if errors.Is(err, dberr.ErrWrongRole) {
				c.PrintErrorf("User with ID %d is not a seller\n", c.sellerId)
				return err
			} else if errors.Is(err, dberr.ErrInvalidPrice) {
				c.PrintErrorf("Invalid price: %d cents\n", c.priceInCents)
				return err
			}

			c.PrintErrorf("Failed to add item to database\n")
			return err
		}
		c.Printf("Item %d added successfully", addedItemId.Int64())

		err = c.printItem(db, addedItemId)
		if err != nil {
			return nil // Don't return an error in this case, as the item is already added to the database.
		}

		return nil
	})
}

func (c *addItemCommand) printItem(db *sql.DB, itemId models.Id) error {
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
