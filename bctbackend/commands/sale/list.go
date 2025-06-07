package sale

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type saleListCommand struct {
	common.Command
}

func NewSaleListCommand() *cobra.Command {
	var command *saleListCommand

	command = &saleListCommand{
		common.Command{
			CobraCommand: &cobra.Command{
				Use:   "list",
				Short: "List all sales",
				Long:  `This command lists all sales in the database.`,
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
			},
		},
	}

	return command.Command.CobraCommand
}

func (command *saleListCommand) execute() error {
	return common.WithOpenedDatabase(command.CobraCommand.ErrOrStderr(), func(db *sql.DB) error {
		saleCount := 0
		tableData := pterm.TableData{
			{"ID", "Cashier", "Transaction Time", "#Items", "Total"},
		}

		addToTable := func(sale *models.SaleSummary) error {
			saleIdString := sale.SaleID.String()
			cashierIdString := sale.CashierID.String()
			transactionTimeString := sale.TransactionTime.FormattedDateTime()
			itemCountString := strconv.Itoa(sale.ItemCount)
			totalString := sale.TotalPriceInCents.DecimalNotation()

			tableData = append(tableData, []string{
				saleIdString,
				cashierIdString,
				transactionTimeString,
				itemCountString,
				totalString,
			})

			saleCount++

			return nil
		}

		if err := queries.GetSales(db, addToTable); err != nil {
			return fmt.Errorf("error while listing sales: %w", err)
		}

		if saleCount == 0 {
			fmt.Println("No sales found")
			return nil
		}

		if err := pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render(); err != nil {
			return fmt.Errorf("error while rendering table: %w", err)
		}

		fmt.Printf("Number of sales listed: %d\n", saleCount)

		return nil
	})
}
