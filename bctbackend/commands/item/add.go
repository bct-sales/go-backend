package item

import (
	"bctbackend/cli/formatting"
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewItemAddCommand() *cobra.Command {
	var description string
	var priceInCents int
	var categoryId int
	var sellerId int
	var donation bool
	var charity bool

	itemAddCommand := cobra.Command{
		Use:   "add",
		Short: "Add an item",
		Long: heredoc.Doc(`
			This command adds a new item to the database.
			The item will always be unfrozen and visible.
			Freezing/hiding needs to be done separately.
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			databasePath := viper.GetString(common.FlagDatabase)

			return common.WithOpenedDatabase(databasePath, func(db *sql.DB) error {
				timestamp := models.Now()

				// We do not want to check the validity for frozen/hidden here, so we just set them to false
				addedItemId, err := queries.AddItem(
					db,
					timestamp,
					description,
					models.MoneyInCents(priceInCents),
					models.Id(categoryId),
					models.Id(sellerId),
					donation,
					charity,
					false,
					false)
				if err != nil {
					return fmt.Errorf("failed to add item to database: %w", err)
				}
				fmt.Println("Item added successfully")

				categoryTable, err := queries.GetCategoryNameTable(db)
				if err != nil {
					slog.Error("An error occurred while trying to get the category map; item is still added to the database", "error", err)
					return nil
				}

				err = formatting.PrintItem(db, categoryTable, addedItemId)
				if err != nil {
					slog.Error("An error occurred while trying to format the output; item is still added to the database", "added item id", addedItemId, "error", err)
					return nil
				}

				return nil
			})
		},
	}

	itemAddCommand.Flags().StringVar(&description, "description", "", "Description of the item")
	itemAddCommand.Flags().IntVar(&priceInCents, "price", 0, "Price of the item in cents")
	itemAddCommand.Flags().IntVar(&categoryId, "category", 0, "ID of the category the item belongs to")
	itemAddCommand.Flags().IntVar(&sellerId, "seller", 0, "ID of the seller of the item")
	itemAddCommand.Flags().BoolVar(&donation, "donation", false, "Whether the item is a donation")
	itemAddCommand.Flags().BoolVar(&charity, "charity", false, "Whether the item is for charity")

	if err := itemAddCommand.MarkFlagRequired("description"); err != nil {
		panic(fmt.Sprintf("failed to mark description flag as required: %v", err))
	}
	if err := itemAddCommand.MarkFlagRequired("price"); err != nil {
		panic(fmt.Sprintf("failed to mark price flag as required: %v", err))
	}
	if err := itemAddCommand.MarkFlagRequired("category"); err != nil {
		panic(fmt.Sprintf("failed to mark category flag as required: %v", err))
	}
	if err := itemAddCommand.MarkFlagRequired("seller"); err != nil {
		panic(fmt.Sprintf("failed to mark seller flag as required: %v", err))
	}

	return &itemAddCommand
}
