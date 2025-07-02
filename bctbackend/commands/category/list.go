package category

import (
	"bctbackend/commands/common"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type listCategoriesCommand struct {
	common.Command
}

func NewCategoryListCommand() *cobra.Command {
	var command *listCategoriesCommand

	command = &listCategoriesCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "list",
				Short: "List all categories",
				Long:  `This command lists all categories in the database.`,
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
			},
		},
	}

	return command.AsCobraCommand()
}

func (c *listCategoriesCommand) execute() error {
	return c.WithOpenedDatabase(func(database *sql.DB) error {
		categories, err := queries.GetCategories(database)
		if err != nil {
			c.PrintErrorf("Failed to list categories\n")
			return fmt.Errorf("failed to list categories: %w", err)
		}

		tableData := pterm.TableData{
			{"ID", "Name"},
		}

		for _, category := range categories {
			categoryIdString := fmt.Sprintf("%d", category.CategoryID)
			categoryNameString := category.Name

			tableData = append(tableData, []string{
				categoryIdString,
				categoryNameString,
			})
		}

		err = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
		if err != nil {
			c.PrintErrorf("Error while rendering table\n")
			return fmt.Errorf("error while rendering table: %w", err)
		}

		return nil
	})
}
