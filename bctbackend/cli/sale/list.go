package sale

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"errors"
	"fmt"
	"strconv"

	"github.com/pterm/pterm"

	_ "modernc.org/sqlite"
)

func ListSales(databasePath string) (r_err error) {
	db, err := database.ConnectToDatabase(databasePath)
	if err != nil {
		return err
	}
	defer func() { r_err = errors.Join(r_err, db.Close()) }()

	saleCount := 0
	tableData := pterm.TableData{
		{"ID", "Cashier", "Transaction Time", "#Items", "Total"},
	}

	addToTable := func(sale *models.SaleSummary) error {
		saleIdString := strconv.FormatInt(sale.SaleId, 10)
		cashierIdString := strconv.FormatInt(sale.CashierId, 10)
		transactionTimeString := models.TimestampToString(sale.TransactionTime)
		itemCountString := strconv.Itoa(sale.ItemCount)
		totalString := strconv.FormatFloat(float64(sale.TotalPriceInCents)/100.0, 'f', 2, 64)

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
		return fmt.Errorf("error while listing sales: %v", err)
	}

	if saleCount == 0 {
		fmt.Println("No sales found")
		return nil
	}

	err = pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render()

	if err != nil {
		return fmt.Errorf("error while rendering table: %v", err)
	}

	fmt.Printf("Number of sales listed: %d\n", saleCount)

	return nil
}
