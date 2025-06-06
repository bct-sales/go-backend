package sale

import (
	"bctbackend/algorithms"
	"bctbackend/cli/formatting"
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

func NewSaleAddCommand() *cobra.Command {
	var rawCashierID uint64
	var rawItemIDs []int64

	command := cobra.Command{
		Use:   "add",
		Short: "Add a new sale",
		Long:  `This command adds a new sale to the database.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return common.WithOpenedDatabase(cmd.ErrOrStderr(), func(db *sql.DB) error {
				timestamp := models.Now()

				itemIDs := algorithms.Map(rawItemIDs, func(id int64) models.Id { return models.Id(id) })
				saleId, err := queries.AddSale(db, models.Id(rawCashierID), timestamp, itemIDs)

				if err != nil {
					return fmt.Errorf("failed to add sale: %w", err)
				}

				fmt.Println("Sale added successfully")

				err = formatting.PrintSale(db, saleId)
				if err != nil {
					slog.Error("An error occurred while trying to format the output; sale is still added to the database.", "error", err)
					return nil // Don't return an error here, as the sale is already added to the database.
				}

				return nil
			})
		},
	}

	command.Flags().Uint64Var(&rawCashierID, "cashier", 0, "ID of the cashier")
	command.Flags().Int64SliceVar(&rawItemIDs, "items", nil, "Items to be added to the sale (comma-separated list of IDs)")
	command.MarkFlagRequired("cashier")
	command.MarkFlagRequired("items")

	return &command
}
