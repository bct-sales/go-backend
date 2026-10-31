package item

import (
	"bctbackend/commands/common"
	"bctbackend/database/queries"
	"context"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type hideItemCommand struct {
	common.Command
}

func NewHideItemCommand() *cobra.Command {
	var command *hideItemCommand

	command = &hideItemCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "hide <item-id> ...",
				Short: "Hides items",
				Long: heredoc.Doc(`
					This command hides items.
					You can unhide them later if needed.
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

func (c *hideItemCommand) execute(ctx context.Context, args []string) error {
	itemIDs, err := c.ParseItemIDs(args)
	if err != nil {
		return err
	}

	transactionErr := c.WithTransaction(ctx, func(ctx context.Context, transaction *queries.TransactionalDatabaseQuerier) error {
		if err := queries.UpdateHiddenStatusOfItems(ctx, transaction, itemIDs, true); err != nil {
			c.PrintErrorf("Failed to hide items: %v\n", err)
			return err
		}

		return nil
	})
	if transactionErr != nil {
		return transactionErr
	}

	c.Printf("Items hidden successfully\n")

	return nil
}
