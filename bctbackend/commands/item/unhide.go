package item

import (
	"bctbackend/commands/common"
	"bctbackend/database/queries"
	"database/sql"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

type unhideItemCommand struct {
	common.Command
}

func NewUnhideItemCommand() *cobra.Command {
	var command *unhideItemCommand

	command = &unhideItemCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "unhide <item-id> ...",
				Short: "Unhides a items",
				Long: heredoc.Doc(`
				This command unhides items.
			   `),
				Args: cobra.MinimumNArgs(1), // Expect exactly one argument (the item ID)
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute(args)
				},
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *unhideItemCommand) execute(args []string) error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		itemIds, err := c.ParseItemIds(args)
		if err != nil {
			return err
		}

		if err := queries.UpdateHiddenStatusOfItems(db, itemIds, false); err != nil {
			c.PrintErrorf("Failed to unhide items: %v\n", err)
			return err
		}

		c.Printf("Items unhidden successfully\n")
		return nil
	})
}
