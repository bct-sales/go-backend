package item

import (
	"bctbackend/cli/formatting"
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

func NewItemCopyCommand() *cobra.Command {
	command := cobra.Command{
		Use:   "copy <item-id>",
		Short: "Copies an item",
		Long: heredoc.Doc(`
			This command makes a copy of an existing item in the database.
			The new item will have the same description, price, category, and seller as the original item.
			The added time will be set to the current time.
			The new item will always be unfrozen and visible.
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.WithOpenedDatabase(cmd.ErrOrStderr(), func(db *sql.DB) error {
				// Parse the item ID from the first argument
				itemId, err := models.ParseId(args[0])
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Invalid item ID: %s\n", args[0])
					return err
				}

				item, err := queries.GetItemWithId(db, itemId)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Failed to retrieve item with ID\n")
					return fmt.Errorf("failed to get item with ID %d: %w", itemId, err)
				}

				timestamp := models.Now()

				copyId, err := queries.AddItem(
					db,
					timestamp,
					item.Description,
					item.PriceInCents,
					item.CategoryID,
					item.SellerID,
					item.Donation,
					item.Charity,
					false,
					false)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Failed to copy item: %v\n", err)
					return fmt.Errorf("failed to insert copy in database: %w", err)
				}

				categoryNameTable, err := queries.GetCategoryNameTable(db)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "An error occurred while trying to get the category map")
					return err
				}

				if err := formatting.PrintItem(db, categoryNameTable, copyId); err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), "Error while printing copied item")
					return fmt.Errorf("failed to print copied item: %w", err)
				}

				return nil
			})
		},
	}

	return &command
}
