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

func NewFreezeItemCommand() *cobra.Command {
	command := cobra.Command{
		Use:   "freeze <item-id>",
		Short: "Freezes an item",
		Long: heredoc.Doc(`
				This command freezes an item so that it cannot be edited anymore.
				You can unfreeze it later if needed.
				Items are automatically frozen when labels are generated for them.
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

				if err := queries.UpdateFreezeStatusOfItems(db, []models.Id{itemId}, true); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Failed to freeze item: %v\n", err)
					return err
				}

				fmt.Fprint(cmd.OutOrStdout(), "Item frozen successfully\n")
				return nil
			})
		},
	}

	return &command
}
