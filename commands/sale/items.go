package sale

import (
	"bctbackend/commands/common"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type ListSoldItemsCommand struct {
	common.Command
}

func NewListSoldItemsCommand() *cobra.Command {
	var command *ListSoldItemsCommand

	command = &ListSoldItemsCommand{
		common.Command{
			CobraCommand: &cobra.Command{
				Use:   "items",
				Short: "List all sold items",
				Long:  `This command lists all sold items.`,
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute()
				},
				Args: cobra.NoArgs,
			},
		},
	}

	return command.AsCobraCommand()
}

func (command *ListSoldItemsCommand) execute() error {
	return command.WithOpenedDatabase(func(db *sql.DB) error {
		categoryNameTable, err := command.GetCategoryNameTable(db)
		if err != nil {
			return err
		}

		tableData := pterm.TableData{
			{
				"Sale ID",
				"Cashier ID",
				"Transaction Time",
				"Item ID",
				"Description",
				"Price",
				"Item Category",
				"Seller ID",
				"Donation",
				"Charity",
			},
		}

		rowCount := 0
		addToTable := func(soldItem *queries.SoldItem) error {
			saleIdStr := soldItem.SaleId.String()
			cashierIdStr := soldItem.CashierId.String()
			transactionTimeStr := soldItem.TransactionTime.FormattedDateTime()
			itemIdStr := soldItem.ItemId.String()
			descriptionStr := soldItem.Description
			priceStr := soldItem.PriceInCents.DecimalNotation()
			itemCategoryStr, ok := categoryNameTable[soldItem.ItemCategory]
			if !ok {
				return fmt.Errorf("unknown category id: %v", soldItem.ItemCategory)
			}
			sellerIdStr := soldItem.SellerId.String()
			donationStr := strconv.FormatBool(soldItem.Donation)
			charityStr := strconv.FormatBool(soldItem.Charity)

			tableData = append(tableData, []string{
				saleIdStr,
				cashierIdStr,
				transactionTimeStr,
				itemIdStr,
				descriptionStr,
				priceStr,
				itemCategoryStr,
				sellerIdStr,
				donationStr,
				charityStr,
			})

			rowCount++

			return nil
		}

		soldItems, err := queries.NewGetSoldItemsQuery().Execute(db)
		if err != nil {
			command.PrintErrorf("Error while listing sales\n")
			return fmt.Errorf("error while listing sales: %w", err)
		}
		for _, soldItem := range soldItems {
			if err := addToTable(soldItem); err != nil {
				command.PrintErrorf("Error while listing sales\n")
				return fmt.Errorf("error while listing sales: %w", err)
			}
		}

		if rowCount == 0 {
			command.Printf("No sold items found\n")
			return nil
		}

		if err := pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render(); err != nil {
			command.PrintErrorf("Error while rendering table\n")
			return fmt.Errorf("error while rendering table: %w", err)
		}

		command.Printf("Number of sales listed: %d\n", rowCount)
		return nil
	})
}
