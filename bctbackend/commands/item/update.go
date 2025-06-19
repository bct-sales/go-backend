package item

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"

	"github.com/MakeNowJust/heredoc"
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
	command.CobraCommand.Flags().BoolVar(&command.charity, "charity", false, "Set item as a charity item")

	command.CobraCommand.MarkFlagRequired("id")

	return command.AsCobraCommand()
}

func (c *updateItemCommand) execute(args []string) error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
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
			c.PrintErrorf("Failed to update item")
			return err
		}

		c.Printf("Item %d updated successfully", c.itemId)
		return nil
	})
}
