package formatting

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"

	"github.com/pterm/pterm"
)

type NoSuchCategoryError struct {
	CategoryId models.Id
}

func (e *NoSuchCategoryError) Error() string {
	return fmt.Sprintf("no category with id %d", e.CategoryId)
}

func PrintUser(user *models.User) error {
	tableData := pterm.TableData{
		{"Property", "Value"},
		{"ID", user.UserId.String()},
		{"Role", user.RoleId.Name()},
		{"Created At", user.CreatedAt.FormattedDateTime()},
		{"Last Activity", FormatOptionalTimestamp(user.LastActivity)},
	}

	err := pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render()
	if err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}

func PrintItems(categoryNameTable map[models.Id]string, items []*models.Item) error {
	tableData := pterm.TableData{
		{"ID", "Description", "Price", "Category", "Seller", "Donation", "Charity", "Added At", "Frozen", "Hidden"},
	}

	for _, item := range items {
		categoryName, ok := categoryNameTable[item.CategoryID]
		if !ok {
			return &NoSuchCategoryError{CategoryId: item.CategoryID}
		}

		tableData = append(tableData, []string{
			item.ItemID.String(),
			item.Description,
			FormatPrice(item.PriceInCents),
			categoryName,
			item.SellerID.String(),
			fmt.Sprintf("%t", item.Donation),
			fmt.Sprintf("%t", item.Charity),
			FormatTimestamp(item.AddedAt),
			fmt.Sprintf("%t", item.Frozen),
			fmt.Sprintf("%t", item.Hidden),
		})
	}

	err := pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render()
	if err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}

func PrintItem(db *sql.DB, categoryNameTable map[models.Id]string, itemId models.Id) error {
	item, err := queries.GetItemWithId(db, itemId)
	if err != nil {
		return fmt.Errorf("failed to get item with id %d: %w", itemId, err)
	}

	categoryName, ok := categoryNameTable[item.CategoryID]
	if !ok {
		return &NoSuchCategoryError{CategoryId: item.CategoryID}
	}

	tableData := pterm.TableData{
		{"Property", "Value"},
		{"Description", item.Description},
		{"Price", FormatPrice(item.PriceInCents)},
		{"Category", categoryName},
		{"Seller", fmt.Sprintf("%d", item.SellerID)},
		{"Donation", fmt.Sprintf("%t", item.Donation)},
		{"Charity", fmt.Sprintf("%t", item.Charity)},
		{"Added At", FormatTimestamp(item.AddedAt)},
	}

	err = pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render()
	if err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}

func PrintSale(db *sql.DB, saleId models.Id) error {
	sale, err := queries.GetSaleWithId(db, saleId)
	if err != nil {
		return fmt.Errorf("failed to get sale with id %d: %w", saleId, err)
	}

	saleItems, err := queries.GetSaleItems(db, saleId)
	if err != nil {
		return fmt.Errorf("failed to get items associated with sale %d: %w", saleId, err)
	}

	tableData := pterm.TableData{
		{"Cashier", sale.CashierID.String()},
		{"Transaction Time", FormatTimestamp(sale.TransactionTime)},
	}

	for index, saleItem := range saleItems {
		tableData = append(tableData, []string{
			fmt.Sprintf("Item %d", index+1),
			saleItem.ItemID.String(),
		})
	}

	if err := pterm.DefaultTable.WithData(tableData).Render(); err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}

func FormatTimestamp(timestamp models.Timestamp) string {
	return timestamp.FormattedDateTime()
}

func FormatOptionalTimestamp(lastActivity *models.Timestamp) string {
	if lastActivity == nil {
		return "N/A"
	}

	return FormatTimestamp(*lastActivity)
}

func FormatRole(roleId models.RoleId) string {
	return roleId.Name()
}

func FormatPrice(priceInCents models.MoneyInCents) string {
	return fmt.Sprintf("€%s", priceInCents.DecimalNotation())
}
