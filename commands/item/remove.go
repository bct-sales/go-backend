package item

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type RemoveItemCommand struct {
	common.Command
}

func NewRemoveItemCommand() *cobra.Command {
	var command *RemoveItemCommand

	command = &RemoveItemCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "remove <item-id> ...",
				Short: "Remove items",
				Long: heredoc.Doc(`
					This command deletes one or more items from the database.
					Note that this is a permanent action and cannot be undone.
					We strongly recommend against using this command unless you are sure you want to delete the item.
					Instead, consider using the 'hide' command to hide the item without deleting it.

					An item cannot be removed if it has been sold.
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

func (c *RemoveItemCommand) execute(args []string) error {
	err := c.WithTransaction(func(db *queries.TransactionalDatabaseQuerier) error {
		for _, arg := range args {
			itemID, err := models.ParseID(arg)
			if err != nil {
				c.PrintErrorf("Invalid item ID: %s\n", args[0])
				return err
			}

			if err := queries.RemoveItemWithID(db, itemID); err != nil {
				c.PrintErrorf("Failed to remove item: %v\n", err)
				return err
			}

			c.Printf("Removed item %d\n", itemID.Int64())
		}

		return nil
	})

	if err != nil {
		c.PrintErrorf("An error occurred while removing the items\nNo items have been removed from the database!\n")
		return err
	}

	c.Printf("All specified items have been successfully removed from the database")
	return nil
}
