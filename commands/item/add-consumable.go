package item

import (
	"bctbackend/commands/common"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type addConsumableCommand struct {
	common.Command
	description  string `exhaustruct:"optional"`
	priceInCents int    `exhaustruct:"optional"`
	categoryId   int    `exhaustruct:"optional"`
	sellerId     int    `exhaustruct:"optional"`
}

func NewAddConsumableCommand() *cobra.Command {
	var command *addConsumableCommand

	command = &addConsumableCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "add-consumable",
				Short: "Add a consumable item",
				Long: heredoc.Doc(`
					Add a consumable item.
				`),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
				Args: cobra.NoArgs,
			},
		},
	}

	command.CobraCommand.Flags().StringVar(&command.description, "description", "", "Description of the item")
	command.CobraCommand.Flags().IntVar(&command.priceInCents, "price", 0, "Price in cents per unit")
	command.CobraCommand.Flags().IntVar(&command.categoryId, "category", 0, "ID of the category the item belongs to")
	command.CobraCommand.Flags().IntVar(&command.sellerId, "seller", 0, "ID of the seller of the item")

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

func (c *addConsumableCommand) execute() error {
	return c.WithTransaction(func(db *queries.TransactionalDatabaseQuerier) error {
		timestamp := models.Now()

		for i := range 5 {
			quantity := i + 1
			description := fmt.Sprintf("%s (%d)", c.description, quantity)
			donation := true
			charity := false
			frozen := false
			hidden := false
			priceInCents := c.priceInCents * quantity

			addedItemId, err := queries.AddItem(
				db,
				timestamp,
				description,
				models.MoneyInCents(priceInCents),
				models.ID(c.categoryId),
				models.ID(c.sellerId),
				donation,
				charity,
				frozen,
				hidden)

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

				c.PrintErrorf("Failed to add item to database\nNo items have been added to the database")
				return err
			}

			c.Printf("Quantity %d -> ID %d\n", quantity, addedItemId.Int64())
		}

		return nil
	})
}
