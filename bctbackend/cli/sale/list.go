package sale

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"
	"strconv"

	"github.com/pterm/pterm"
)

func ListSales(databasePath string) (r_err error) {
	db, err := database.OpenDatabase(databasePath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	saleCount := 0
	tableData := pterm.TableData{
		{"ID", "Cashier", "Transaction Time", "#Items", "Total"},
	}

	addToTable := func(sale *models.SaleSummary) error {
		saleIdString := sale.SaleId.String()
		cashierIdString := sale.CashierId.String()
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

	err = pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render()
	if err != nil {
		return fmt.Errorf("error while rendering table: %w", err)
	}

	fmt.Printf("Number of sales listed: %d\n", saleCount)

	return nil
}
