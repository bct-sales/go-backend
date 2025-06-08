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

func NewItemAddCommand() *cobra.Command {
	var description string
	var priceInCents int
	var categoryId int
	var sellerId int
	var donation bool
	var charity bool

	command := cobra.Command{
		Use:   "add",
		Short: "Add an item",
		Long: heredoc.Doc(`
			This command adds a new item to the database.
			The item will always be unfrozen and visible.
			Freezing/hiding needs to be done separately.
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.WithOpenedDatabase(cmd.ErrOrStderr(), func(db *sql.DB) error {
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
					fmt.Fprintf(cmd.ErrOrStderr(), "Failed to add item to database: %v\n", err)
					return err
				}
				fmt.Println("Item added successfully")

				categoryNameTable, err := queries.GetCategoryNameTable(db)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "An error occurred while trying to get the category map: %v\n", err)
					return err
				}

				err = formatting.PrintItem(db, categoryNameTable, addedItemId)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "An error occurred while trying to format the output: %v\n", err)
					return err
				}

				return nil
			})
		},
	}

	command.Flags().StringVar(&description, "description", "", "Description of the item")
	command.Flags().IntVar(&priceInCents, "price", 0, "Price of the item in cents")
	command.Flags().IntVar(&categoryId, "category", 0, "ID of the category the item belongs to")
	command.Flags().IntVar(&sellerId, "seller", 0, "ID of the seller of the item")
	command.Flags().BoolVar(&donation, "donation", false, "Whether the item is a donation")
	command.Flags().BoolVar(&charity, "charity", false, "Whether the item is for charity")

	if err := command.MarkFlagRequired("description"); err != nil {
		panic(fmt.Sprintf("failed to mark description flag as required: %v", err))
	}
	if err := command.MarkFlagRequired("price"); err != nil {
		panic(fmt.Sprintf("failed to mark price flag as required: %v", err))
	}
	if err := command.MarkFlagRequired("category"); err != nil {
		panic(fmt.Sprintf("failed to mark category flag as required: %v", err))
	}
	if err := command.MarkFlagRequired("seller"); err != nil {
		panic(fmt.Sprintf("failed to mark seller flag as required: %v", err))
	}

	return &command
}
