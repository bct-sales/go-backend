package formatting

import (
	"bctbackend/database/models"
	"fmt"
	"strconv"

	"github.com/pterm/pterm"
)

type NoSuchCategoryError struct {
	CategoryId models.Id
}

func (e *NoSuchCategoryError) Error() string {
	return fmt.Sprintf("no category with id %d", e.CategoryId)
}

func PrintUser(user *models.User) error {
	var lastActivityString string
	if user.LastActivity != nil {
		lastActivityString = user.LastActivity.FormattedDateTime()
	} else {
		lastActivityString = "never"
	}

	tableData := pterm.TableData{
		{"Property", "Value"},
		{"ID", user.UserId.String()},
		{"Role", user.RoleId.Name()},
		{"Created At", user.CreatedAt.FormattedDateTime()},
		{"Last Activity", lastActivityString},
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
			item.PriceInCents.DecimalNotation(),
			categoryName,
			item.SellerID.String(),
			strconv.FormatBool(item.Donation),
			strconv.FormatBool(item.Charity),
			item.AddedAt.FormattedDateTime(),
			strconv.FormatBool(item.Frozen),
			strconv.FormatBool(item.Hidden),
		})
	}

	err := pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render()
	if err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}
