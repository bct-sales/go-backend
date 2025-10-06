package item

import (
	"bctbackend/commands/common"
	"bctbackend/database/queries"

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
					return command.execute(args)
				},
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *hideItemCommand) execute(args []string) error {
	itemIds, err := c.ParseItemIDs(args)
	if err != nil {
		return err
	}

	transactionErr := c.WithTransaction(func(transaction *queries.TransactionalDatabaseQuerier) error {
		if err := queries.UpdateHiddenStatusOfItems(transaction, itemIds, true); err != nil {
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
