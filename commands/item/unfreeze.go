package item

import (
	"bctbackend/commands/common"
	"bctbackend/database/queries"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type unfreezeItemCommand struct {
	common.Command
}

func NewUnfreezeItemCommand() *cobra.Command {
	var command *unfreezeItemCommand

	command = &unfreezeItemCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "unfreeze <item-id> ...",
				Short: "Unfreezes items",
				Long: heredoc.Doc(`
					This command unfreezes items so that they can be edited again.
					Use with care: if labels were printed for the item, editing the item would make them inaccurate.
					It is highly recommended not to unfreeze items but instead create a new item with the updated information.
					See the item copy command.
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

func (c *unfreezeItemCommand) execute(args []string) error {
	itemIds, err := c.ParseItemIDs(args)
	if err != nil {
		return err
	}

	transactionErr := c.WithTransaction(func(db *queries.TransactionalDatabaseQuerier) error {
		if err := queries.UpdateFreezeStatusOfItems(db, itemIds, false); err != nil {
			c.PrintErrorf("Failed to unfreeze items: %v\n", err)
			return err
		}

		return nil
	})
	if transactionErr != nil {
		return transactionErr
	}

	c.Printf("Items unfrozen successfully\n")
	return nil
}
