package item

import (
	"bctbackend/commands/common"
	"bctbackend/database/queries"
	"database/sql"

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
				Short: "Unfreezes an item",
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
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		itemIds, err := c.ParseItemIds(args)
		if err != nil {
			return err
		}

		if err := queries.UpdateFreezeStatusOfItems(db, itemIds, false); err != nil {
			c.PrintErrorf("Failed to unfreeze item: %v\n", err)
			return err
		}

		c.Printf("Items unfrozen successfully\n")
		return nil
	})
}
