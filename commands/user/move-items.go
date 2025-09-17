package user

import (
	"bctbackend/commands/common"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type moveItemsCommand struct {
	common.Command
	oldSeller    uint64
	newSeller    uint64
	forceFrozen  bool
	forceReceive bool
}

func NewMoveItemsCommand() *cobra.Command {
	var command *moveItemsCommand

	command = &moveItemsCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "move-items",
				Short: "Move items from one seller to another",
				Long: heredoc.Doc(`
					This command moves all items owner by oldSeller
					to newSeller.
				`),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
				Args: cobra.NoArgs,
			},
		},
	}

	command.CobraCommand.Flags().Uint64Var(&command.oldSeller, "from", 0, "Id of old seller")
	command.CobraCommand.Flags().Uint64Var(&command.newSeller, "to", 0, "Id of new seller")
	command.CobraCommand.Flags().BoolVar(&command.forceFrozen, "force-frozen", false, "Force frozen items")
	command.CobraCommand.Flags().BoolVar(&command.forceReceive, "force-merge", false, "Force even if receiver seller already has items")
	command.CobraCommand.MarkFlagRequired("from")
	command.CobraCommand.MarkFlagRequired("to")

	return command.AsCobraCommand()
}

func (c *moveItemsCommand) execute() error {
	return c.WithTransaction(func(transaction *queries.TransactionalDatabaseQuerier) error {
		oldSeller := models.Id(c.oldSeller)
		newSeller := models.Id(c.newSeller)

		// Check if donating seller exists and is indeed a seller
		if err := queries.EnsureUserExistsAndHasRole(transaction, oldSeller, models.NewSellerRoleId()); err != nil {
			if errors.Is(err, dberr.ErrNoSuchUser) {
				c.PrintErrorf("User %s does not exist\n", oldSeller.String())
				return err
			}

			if errors.Is(err, dberr.ErrWrongRole) {
				c.PrintErrorf("User %s is not a seller\n", oldSeller.String())
				return err
			}

			c.PrintErrorf("An error occurred while checking the validity of the old seller %s: %v\n", oldSeller.String(), err)
			return err
		}

		// Check if receiving seller exists and is indeed a seller
		if err := queries.EnsureUserExistsAndHasRole(transaction, newSeller, models.NewSellerRoleId()); err != nil {
			if errors.Is(err, dberr.ErrNoSuchUser) {
				c.PrintErrorf("User %s does not exist\n", newSeller.String())
				return err
			}

			if errors.Is(err, dberr.ErrWrongRole) {
				c.PrintErrorf("User %s is not a seller\n", newSeller.String())
				return err
			}

			c.PrintErrorf("An error occurred while checking the validity of the new seller %s: %v\n", newSeller.String(), err)
			return err
		}

		// Check if any of the items is frozen
		hasFrozen, hasFrozenError := queries.DoesSellerHaveFrozenItems(transaction, oldSeller)
		if hasFrozenError != nil {
			c.PrintErrorf("An error occured while checking for frozen items: %v\n", hasFrozenError)
			return hasFrozenError
		}
		if hasFrozen && !c.forceFrozen {
			c.PrintErrorf("Seller %s has frozen items\nChanging the items' owner would invalidate labels\nUse --force-frozen to move items anyway\n", oldSeller.String())
			return dberr.ErrItemFrozen
		}

		// Check if the receiving seller already has items
		newSellerItemCount, newSellerItemCountError := queries.CountSellerItems(transaction, newSeller, queries.IncludeAll, queries.IncludeAll)
		if newSellerItemCountError != nil {
			c.PrintErrorf("An error occurred while counting the receiving seller's items: %v", newSellerItemCountError)
			return newSellerItemCountError
		}
		if newSellerItemCount > 0 && !c.forceReceive {
			c.PrintErrorf("Seller %s already has items, namely %d\nAre you sure you want to merge sellers?\nUse --force-merge to move items anyway\n", newSeller.String(), newSellerItemCount)
			return &ErrReceiverHasItems{}
		}

		// Perform move
		if err := queries.MoveItemsToNewSeller(transaction, oldSeller, newSeller); err != nil {
			c.PrintErrorf("Error moving items from seller %s to seller %s: %v\n", oldSeller.String(), newSeller.String(), err)
			return err
		}

		c.Printf("Successfully moved items")

		return nil
	})
}

type ErrReceiverHasItems struct{}

func (e *ErrReceiverHasItems) Error() string {
	return "receiver already has items"
}
