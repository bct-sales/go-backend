package category

import (
	"bctbackend/commands/common"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/urfave/cli/v2"
)

type categoryCountCommand struct {
	common.Command
	includeHidden bool
}

func NewCategoryCountCommand() *cobra.Command {
	var command *categoryCountCommand

	command = &categoryCountCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "count",
				Short: "Item counts by category",
				Long:  `This command the number of items in each category.`,
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
			},
		},
	}

	command.CobraCommand.Flags().BoolVar(&command.includeHidden, "include-hidden", false, "Include hidden items in the count")

	return command.AsCobraCommand()
}

func (c *categoryCountCommand) execute() error {
	return c.WithOpenedDatabase(func(database *sql.DB) error {
		var itemSelection queries.ItemSelection
		if c.includeHidden {
			itemSelection = queries.AllItems
		} else {
			itemSelection = queries.OnlyVisibleItems
		}

		categoryCounts, err := queries.GetCategoryCounts(database, itemSelection)
		if err != nil {
			return fmt.Errorf("failed to get category counts: %w", err)
		}

		categoryTable, err := queries.GetCategoryNameTable(database)
		if err != nil {
			return fmt.Errorf("failed to get category name table: %w", err)
		}

		tableData := pterm.TableData{
			{"ID", "Name", "Count"},
		}

		for categoryId, categoryCount := range categoryCounts {
			categoryNameString, ok := categoryTable[categoryId]
			if !ok {
				return cli.Exit(fmt.Sprintf("Bug: unknown category %d", categoryId), 1)
			}
			categoryIdString := categoryId.String()
			count := strconv.Itoa(categoryCount)

			tableData = append(tableData, []string{
				categoryIdString,
				categoryNameString,
				count,
			})
		}

		if err = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render(); err != nil {
			c.PrintErrorf("Error while rendering table")
			return fmt.Errorf("error while rendering table: %w", err)
		}

		return nil
	})
}
