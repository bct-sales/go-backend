package item

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func NewItemUnfreezeCommand() *cobra.Command {
	itemListCommand := cobra.Command{
		Use:   "unfreeze ID",
		Short: "Unfreezes an item",
		Long: heredoc.Doc(`
				This command unfreezes an item so that it can be edited again.
				Use with care: if labels were printed for the item, editing the item would make them inaccurate.
				It is highly recommended to not unfreeze items and instead create a new item with the updated information.
				See the item copy command.
			   `),
		Args: cobra.ExactArgs(1), // Expect exactly one argument (the item ID)
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.WithOpenedDatabase(cmd.ErrOrStderr(), func(db *sql.DB) error {
				// Parse the item ID from the first argument
				itemId, err := models.ParseId(args[0])
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Invalid item ID: %s\n", args[0])
					return err
				}

				if err := queries.UpdateFreezeStatusOfItems(db, []models.Id{itemId}, false); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Failed to unfreeze item: %v\n", err)
					return err
				}

				fmt.Println("Item unfrozen successfully")
				return nil
			})
		},
	}

	return &itemListCommand
}
