package item

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"

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
				Use:   "remove <item-id>",
				Short: "Removes an item",
				Long: heredoc.Doc(`
				This command deletes an item from the database.
				Note that this is a permanent action and cannot be undone.
				We strongly recommend against using this command unless you are sure you want to delete the item.
				Instead, consider using the 'hide' command to hide the item without deleting it.

				An item cannot be removed if it has been sold.
			   `),
				Args: cobra.ExactArgs(1), // Expect exactly one argument (the item ID)
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute(args)
				},
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *RemoveItemCommand) execute(args []string) error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		// Parse the item ID from the first argument
		itemId, err := models.ParseId(args[0])
		if err != nil {
			c.PrintErrorf("Invalid item ID: %s\n", args[0])
			return err
		}

		if err := queries.RemoveItemWithId(db, itemId); err != nil {
			c.PrintErrorf("Failed to remove item: %v\n", err)
			return err
		}

		c.Printf("Item removed successfully\n")
		return nil
	})
}
