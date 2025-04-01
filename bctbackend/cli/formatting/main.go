package formatting

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/defs"
	"database/sql"
	"fmt"

	"github.com/pterm/pterm"
)

func PrintUser(user *models.User) error {
	tableData := pterm.TableData{
		{"Property", "Value"},
		{"ID", FormatId(user.UserId)},
		{"Role", FormatRole(user.RoleId)},
		{"Created At", FormatTimestamp(user.CreatedAt)},
		{"Last Activity", FormatOptionalTimestamp(user.LastActivity)},
	}

	err := pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render()

	if err != nil {
		return err
	}

	return nil
}

func PrintItems(items []*models.Item) error {
	tableData := pterm.TableData{
		{"ID", "Description", "Price", "Category", "Donation", "Charity", "Added At"},
	}

	for _, item := range items {
		categoryName, err := defs.NameOfCategory(item.CategoryId)

		if err != nil {
			return err
		}

		tableData = append(tableData, []string{
			FormatId(item.ItemId),
			item.Description,
			FormatPrice(item.PriceInCents),
			categoryName,
			fmt.Sprintf("%t", item.Donation),
			fmt.Sprintf("%t", item.Charity),
			FormatTimestamp(item.AddedAt),
		})
	}

	err := pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render()

	if err != nil {
		return err
	}

	return nil
}

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
		{"Price", FormatPrice(item.PriceInCents)},
		{"Category", categoryName},
		{"Seller", fmt.Sprintf("%d", item.SellerId)},
		{"Donation", fmt.Sprintf("%t", item.Donation)},
		{"Charity", fmt.Sprintf("%t", item.Charity)},
		{"Added At", FormatTimestamp(item.AddedAt)},
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
		{"Transaction Time", FormatTimestamp(sale.TransactionTime)},
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
	return models.TimestampToString(timestamp)
}

func FormatOptionalTimestamp(lastActivity *models.Timestamp) string {
	if lastActivity == nil {
		return "N/A"
	}

	return FormatTimestamp(*lastActivity)
}

func FormatRole(roleId models.Id) string {
	string, err := models.NameOfRole(roleId)

	if err != nil {
		return fmt.Sprintf("<error: unknown role %d>", roleId)
	}

	return string
}

func FormatPrice(priceInCents models.MoneyInCents) string {
	return fmt.Sprintf("$%.2f", float64(priceInCents)/100.0)
}
