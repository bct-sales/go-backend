package category

import (
	"bctbackend/commands/common"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"
	"maps"
	"slices"
	"strconv"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
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
		categoryCounts, err := c.getCategoryCounts(database)
		if err != nil {
			return err
		}

		categoryNameTable, err := c.GetCategoryNameTable(database)
		if err != nil {
			return err
		}

		if err := c.printCategoryCounts(categoryCounts, categoryNameTable); err != nil {
			return err
		}

		return nil
	})
}

func (c *categoryCountCommand) getCategoryCounts(database *sql.DB) (map[models.Id]int, error) {
	var itemSelection queries.ItemSelection

	if c.includeHidden {
		itemSelection = queries.AllItems
	} else {
		itemSelection = queries.OnlyVisibleItems
	}

	categoryCounts, err := queries.GetCategoryCounts(database, itemSelection)

	if err != nil {
		c.PrintErrorf("Failed to get category counts")
		return nil, fmt.Errorf("failed to get category counts: %w", err)
	}

	return categoryCounts, nil
}

func (c *categoryCountCommand) printCategoryCounts(categoryCounts map[models.Id]int, categoryNameTable map[models.Id]string) error {
	tableData := pterm.TableData{
		{"ID", "Name", "Count"},
	}

	categoryIds := maps.Keys(categoryCounts)
	sortedCategoryIds := slices.Sorted(categoryIds)

	for _, categoryId := range sortedCategoryIds {
		categoryCount, ok := categoryCounts[categoryId]
		if !ok {
			panic("Bug: category ID not found in counts map")
		}

		categoryName, ok := categoryNameTable[categoryId]
		if !ok {
			panic("Bug: category ID not found in category name table")
		}

		categoryIdString := categoryId.String()
		countString := strconv.Itoa(categoryCount)
		tableData = append(tableData, []string{
			categoryIdString,
			categoryName,
			countString,
		})
	}

	if err := pterm.DefaultTable.WithHasHeader().WithData(tableData).Render(); err != nil {
		c.PrintErrorf("Error while rendering table")
		return fmt.Errorf("error while rendering table: %w", err)
	}

	return nil
}
