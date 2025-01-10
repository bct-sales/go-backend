package formatting

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/defs"
	"database/sql"
	"fmt"
	"time"

	"github.com/pterm/pterm"
)

func PrintItem(db *sql.DB, itemId models.Id) error {
	item, err := queries.GetItemWithId(db, itemId)

	if err != nil {
		return err
	}

	categoryName, err := defs.NameOfCategory(item.CategoryId)

	if err != nil {
		return err
	}

	tableData := pterm.TableData{
		{"Property", "Value"},
		{"Description", item.Description},
		{"Price", fmt.Sprintf("%.2f", float64(item.PriceInCents)/100.0)},
		{"Category", categoryName},
		{"Seller", fmt.Sprintf("%d", item.SellerId)},
		{"Donation", fmt.Sprintf("%t", item.Donation)},
		{"Charity", fmt.Sprintf("%t", item.Charity)},
	}

	err = pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render()

	if err != nil {
		return err
	}

	return nil
}

func PrintSale(db *sql.DB, saleId models.Id) error {
	sale, err := queries.GetSaleWithId(db, saleId)

	if err != nil {
		return err
	}

	saleItems, err := queries.GetSaleItems(db, saleId)

	if err != nil {
		return err
	}

	tableData := pterm.TableData{
		{"Cashier", FormatId(sale.CashierId)},
		{"Transaction Time", fmt.Sprintf("%d", sale.TransactionTime)},
	}

	for index, saleItem := range saleItems {
		tableData = append(tableData, []string{
			fmt.Sprintf("Item %d", index+1),
			FormatId(saleItem.ItemId),
		})
	}

	err = pterm.DefaultTable.WithData(tableData).Render()

	if err != nil {
		return err
	}

	return nil
}

func FormatId(id models.Id) string {
	return fmt.Sprintf("%d", id)
}

func FormatTimestamp(timestamp models.Timestamp) string {
	return time.Unix(timestamp, 0).String()
}
