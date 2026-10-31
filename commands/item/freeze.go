package item

import (
	"bctbackend/commands/common"
	"bctbackend/database/queries"
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type freezeItemCommand struct {
	common.Command
}

func NewFreezeItemCommand() *cobra.Command {
	var command *freezeItemCommand

	command = &freezeItemCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "freeze <item-id> ...",
				Short: "Freezes items",
				Long: heredoc.Doc(`
					This command freezes items so that they cannot be edited anymore.
					You can unfreeze them later if needed.
					Items are automatically frozen when labels are generated for them.
			   `),
				Args: cobra.MinimumNArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute(cmd.Context(), args)
				},
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *freezeItemCommand) execute(ctx context.Context, args []string) error {
	itemIDs, err := c.ParseItemIDs(args)
	if err != nil {
		return err
	}

	transactionErr := c.WithTransaction(ctx, func(ctx context.Context, transaction *queries.TransactionalDatabaseQuerier) error {
		if err := queries.UpdateFreezeStatusOfItems(ctx, transaction, itemIDs, true); err != nil {
			c.PrintErrorf("Failed to freeze items: %v\n", err)
			return err
		}

		return nil
	})
	if transactionErr != nil {
		return transactionErr
	}

	c.Printf("Items frozen successfully\n")
	return nil
}
