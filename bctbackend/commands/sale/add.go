package sale

import (
	"bctbackend/algorithms"
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"

	"github.com/pterm/pterm"
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
	if err := c.EnsureConfigurationFileLoaded(); err != nil {
		return err
	}

	return c.WithOpenedDatabase(func(db *sql.DB) error {
		timestamp := models.Now()

		itemIDs := algorithms.Map(c.rawItemIDs, func(id int64) models.Id { return models.Id(id) })
		saleId, err := queries.AddSale(db, models.Id(c.rawCashierID), timestamp, itemIDs)

		if err != nil {
			return fmt.Errorf("failed to add sale: %w", err)
		}

		c.Printf("Sale added successfully\n")

		err = c.printSale(db, saleId)
		if err != nil {
			return nil // Don't return an error in this case, as the sale is already added to the database.
		}

		return nil
	})
}

func (c *addNewSaleCommand) printSale(db *sql.DB, saleId models.Id) error {
	sale, err := queries.GetSaleWithId(db, saleId)
	if err != nil {
		c.PrintErrorf("Failed to get sale back from database")
		return fmt.Errorf("failed to get sale with id %d: %w", saleId, err)
	}

	saleItems, err := queries.GetSaleItems(db, saleId)
	if err != nil {
		c.PrintErrorf("Failed to get items associated with sale")
		return fmt.Errorf("failed to get items associated with sale %d: %w", saleId, err)
	}

	tableData := pterm.TableData{
		{"Cashier", sale.CashierID.String()},
		{"Transaction Time", sale.TransactionTime.FormattedDateTime()},
	}

	for index, saleItem := range saleItems {
		tableData = append(tableData, []string{
			fmt.Sprintf("Item %d", index+1),
			saleItem.ItemID.String(),
		})
	}

	if err := pterm.DefaultTable.WithData(tableData).Render(); err != nil {
		c.PrintErrorf("Failed to render sale table")
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}
