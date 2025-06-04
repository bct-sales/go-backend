package item

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewItemFreezeCommand() *cobra.Command {
	itemListCommand := cobra.Command{
		Use:   "freeze ID",
		Short: "Freezes item",
		Long: heredoc.Doc(`
				This command freezes an item so that it cannot be edited anymore.
				You can unfreeze it later if needed.
				Items are automatically frozen when labels are generated for them.
			   `),
		Args: cobra.ExactArgs(1), // Expect exactly one argument (the item ID)
		Run: func(cmd *cobra.Command, args []string) {
			databasePath := viper.GetString(common.FlagDatabase)

			err := common.WithOpenedDatabase(databasePath, func(db *sql.DB) error {
				// Parse the item ID from the first argument
				itemId, err := models.ParseId(args[0])
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Invalid item ID: %s\n", args[0])
				}

				if err := queries.UpdateFreezeStatusOfItems(db, []models.Id{itemId}, true); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Failed to freeze item: %v\n", err)
					return nil
				}

				fmt.Println("Item frozen successfully")
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
