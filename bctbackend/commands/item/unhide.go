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

func NewUnhideItemCommand() *cobra.Command {
	command := cobra.Command{
		Use:   "unhide <item-id>",
		Short: "Unhides an item",
		Long: heredoc.Doc(`
				This command unhides an item.
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

				if err := queries.UpdateHiddenStatusOfItems(db, []models.Id{itemId}, false); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Failed to unhide item: %v\n", err)
					return fmt.Errorf("failed to update database: %w", err)
				}

				fmt.Println("Item unhidden successfully")
				return nil
			})
		},
	}

	return &command
}
