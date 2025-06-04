package item

import (
	"bctbackend/cli/formatting"
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewItemShowCommand() *cobra.Command {
	itemListCommand := cobra.Command{
		Use:   "show ID",
		Short: "Show item info",
		Long:  `This command shows detailed information about a specific item.`,
		Args:  cobra.ExactArgs(1), // Expect exactly one argument (the item ID)
		Run: func(cmd *cobra.Command, args []string) {
			databasePath := viper.GetString(common.FlagDatabase)

			err := common.WithOpenedDatabase(databasePath, func(db *sql.DB) error {
				// Parse the item ID from the first argument
				itemId, err := models.ParseId(args[0])
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Invalid item ID: %s\n", args[0])
				}

				// We need the category table to print user-friendly category names
				categoryTable, err := queries.GetCategoryNameTable(db)
				if err != nil {
					return fmt.Errorf("failed to get category name table: %w", err)
				}

				// Print the item details
				err = formatting.PrintItem(db, categoryTable, itemId)
				if err != nil {
					return fmt.Errorf("failed to print item: %w", err)
				}

				return nil
			})

			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Error while showing item: %v\n", err)
				return
			}
		},
	}

	return &itemListCommand
}
