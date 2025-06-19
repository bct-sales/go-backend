package item

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
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
				Use:   "unfreeze <item-id>",
				Short: "Unfreezes an item",
				Long: heredoc.Doc(`
				This command unfreezes an item so that it can be edited again.
				Use with care: if labels were printed for the item, editing the item would make them inaccurate.
				It is highly recommended to not unfreeze items and instead create a new item with the updated information.
				See the item copy command.
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

func (c *unfreezeItemCommand) execute(args []string) error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		itemId, err := models.ParseId(args[0])
		if err != nil {
			c.PrintErrorf("Invalid item ID: %s\n", args[0])
			return err
		}

		if err := queries.UpdateFreezeStatusOfItems(db, []models.Id{itemId}, false); err != nil {
			c.PrintErrorf("Failed to unfreeze item: %v\n", err)
			return err
		}

		c.Printf("Item unfrozen successfully\n")
		return nil
	})
}
