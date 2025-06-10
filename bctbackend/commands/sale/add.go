package sale

import (
	"bctbackend/algorithms"
	"bctbackend/cli/formatting"
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"

	"github.com/spf13/cobra"
)

type addNewSaleCommand struct {
	common.Command
	rawCashierID uint64
	rawItemIDs   []int64
}

func NewSaleAddCommand() *cobra.Command {
	var command *addNewSaleCommand

	command = &addNewSaleCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "add",
				Short: "Add a new sale",
				Long:  `This command adds a new sale to the database.`,
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
			},
		},
	}

	command.CobraCommand.Flags().Uint64Var(&command.rawCashierID, "cashier", 0, "ID of the cashier")
	command.CobraCommand.Flags().Int64SliceVar(&command.rawItemIDs, "items", nil, "Items to be added to the sale (comma-separated list of IDs)")
	command.CobraCommand.MarkFlagRequired("cashier")
	command.CobraCommand.MarkFlagRequired("items")

	return command.AsCobraCommand()
}

func (c *addNewSaleCommand) execute() error {
	return c.WithOpenedDatabase(func(db *sql.DB) error {
		timestamp := models.Now()

		itemIDs := algorithms.Map(c.rawItemIDs, func(id int64) models.Id { return models.Id(id) })
		saleId, err := queries.AddSale(db, models.Id(c.rawCashierID), timestamp, itemIDs)

		if err != nil {
			return fmt.Errorf("failed to add sale: %w", err)
		}

		c.Printf("Sale added successfully")

		err = formatting.PrintSale(db, saleId)
		if err != nil {
			c.PrintErrorf("Failed to show resulting sale; sale has been successfully added to the database though")
			return nil // Don't return an error here, as the sale is already added to the database.
		}

		return nil
	})
}
