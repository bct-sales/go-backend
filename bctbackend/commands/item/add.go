package item

import (
	"bctbackend/cli/formatting"
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"

	"github.com/MakeNowJust/heredoc"
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
			c.PrintErrorf("Failed to add item to database\n")
			return err
		}
		c.Printf("Item added successfully")

		categoryNameTable, err := c.GetCategoryNameTable(db)
		if err != nil {
			return err
		}

		err = formatting.PrintItem(db, categoryNameTable, addedItemId)
		if err != nil {
			c.PrintErrorf("An error occurred while trying to format the output\n")
			return err
		}

		return nil
	})
}
