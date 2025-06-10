package formatting

import (
	"bctbackend/database/models"
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
