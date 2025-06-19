package item

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"

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
				Use:   "freeze <item-id>",
				Short: "Freezes an item",
				Long: heredoc.Doc(`
				This command freezes an item so that it cannot be edited anymore.
				You can unfreeze it later if needed.
				Items are automatically frozen when labels are generated for them.
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

func (c *freezeItemCommand) execute(args []string) error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		itemId, err := models.ParseId(args[0])
		if err != nil {
			c.PrintErrorf("Invalid item ID: %s\n", args[0])
			return err
		}

		if err := queries.UpdateFreezeStatusOfItems(db, []models.Id{itemId}, true); err != nil {
			c.PrintErrorf("Failed to freeze item: %v\n", err)
			return err
		}

		c.Printf("Item frozen successfully\n")
		return nil
	})
}
