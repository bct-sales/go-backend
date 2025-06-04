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

func NewItemListCommand() *cobra.Command {
	var showHidden bool

	itemListCommand := cobra.Command{
		Use:   "list",
		Short: "List all items",
		Long:  `This command lists all items in the database.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			databasePath := viper.GetString(common.FlagDatabase)

			return common.WithOpenedDatabase(databasePath, func(db *sql.DB) error {
				var hiddenStrategy queries.ItemSelection
				if showHidden {
					hiddenStrategy = queries.AllItems
				} else {
					hiddenStrategy = queries.OnlyVisibleItems
				}

				items := []*models.Item{}
				if err := queries.GetItems(db, queries.CollectTo(&items), hiddenStrategy); err != nil {
					return fmt.Errorf("error while listing items: %w", err)
				}

				itemCount := len(items)

				if itemCount > 0 {
					categoryTable, err := queries.GetCategoryNameTable(db)
					if categoryTable == nil {
						return fmt.Errorf("error while getting category map: %w", err)
					}

					if err := formatting.PrintItems(categoryTable, items); err != nil {
						return fmt.Errorf("error while rendering table: %w", err)
					}

					fmt.Printf("Number of items listed: %d\n", itemCount)

					return nil
				} else {
					fmt.Println("No items found")

					return nil
				}
			})
		},
	}

	itemListCommand.Flags().BoolVar(&showHidden, "show-hidden", false, "Show hidden items")

	return &itemListCommand
}
