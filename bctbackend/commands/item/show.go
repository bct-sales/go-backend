package item

import (
	"bctbackend/cli/formatting"
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
)

func NewItemShowCommand() *cobra.Command {
	command := cobra.Command{
		Use:   "show ID",
		Short: "Show item info",
		Long:  `This command shows detailed information about a specific item.`,
		Args:  cobra.ExactArgs(1), // Expect exactly one argument (the item ID)
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.WithOpenedDatabase(cmd.ErrOrStderr(), func(db *sql.DB) error {
				// Parse the item ID from the first argument
				itemId, err := models.ParseId(args[0])
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Invalid item ID: %s\n", args[0])
					return err
				}

				// We need the category table to print user-friendly category names
				categoryTable, err := queries.GetCategoryNameTable(db)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "An error occurred while trying to get the category map: %v\n", err)
					return err
				}

				// Print the item details
				err = formatting.PrintItem(db, categoryTable, itemId)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error while printing item: %v\n", err)
					return err
				}

				return nil
			})
		},
	}

	return &command
}
