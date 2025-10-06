package sale

import (
	"bctbackend/commands/common"
	dbcsv "bctbackend/database/csv"
	"bctbackend/database/queries"
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type ListSoldItemsCommand struct {
	common.Command
	format string
}

func NewListSoldItemsCommand() *cobra.Command {
	var command *ListSoldItemsCommand

	command = &ListSoldItemsCommand{
		Command: common.Command{
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
		format: "table",
	}

	command.CobraCommand.Flags().StringVar(&command.format, "format", "table", "Output format (format, csv)")

	return command.AsCobraCommand()
}

func (command *ListSoldItemsCommand) execute() error {
	return command.WithOpenedDatabase(func(db *sql.DB) error {
		switch command.format {
		case "table":
			return command.listSoldItemsInTableFormat(db)
		case "csv":
			return command.listSoldItemsInCsvFormat(db)
		default:
			command.PrintErrorf("Invalid format: %s\n", command.format)
			return fmt.Errorf("unknown format: %s", command.format)
		}
	})
}

func (command *ListSoldItemsCommand) listSoldItemsInCsvFormat(db *sql.DB) error {
	categoryNameTable, err := command.GetCategoryNameTable(db)
	if err != nil {
		return err
	}

	soldItems, err := queries.NewGetSoldItemsQuery().Execute(db)
	if err != nil {
		command.PrintErrorf("Error while listing sales\n")
		return fmt.Errorf("error while listing sales: %w", err)
	}

	dbcsv.FormatSoldItemsAsCSV(soldItems, categoryNameTable, os.Stdout)
	if err != nil {
		command.PrintErrorf("Error while formatting sold items as CSV\n")
		return fmt.Errorf("failed to format sold items as a CSV: %w", err)
	}

	return nil
}

func (command *ListSoldItemsCommand) listSoldItemsInTableFormat(db *sql.DB) error {
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
		saleIDStr := soldItem.SaleID.String()
		cashierIDStr := soldItem.CashierID.String()
		transactionTimeStr := soldItem.TransactionTime.FormattedDateTime()
		itemIDStr := soldItem.ItemID.String()
		descriptionStr := soldItem.Description
		priceStr := soldItem.PriceInCents.DecimalNotation()
		itemCategoryStr, ok := categoryNameTable[soldItem.ItemCategoryID]
		if !ok {
			return fmt.Errorf("unknown category id: %v", soldItem.ItemCategoryID)
		}
		sellerIDStr := soldItem.SellerId.String()
		donationStr := strconv.FormatBool(soldItem.Donation)
		charityStr := strconv.FormatBool(soldItem.Charity)

		tableData = append(tableData, []string{
			saleIDStr,
			cashierIDStr,
			transactionTimeStr,
			itemIDStr,
			descriptionStr,
			priceStr,
			itemCategoryStr,
			sellerIDStr,
			donationStr,
			charityStr,
		})

		rowCount++

		return nil
	}

	soldItems, err := queries.NewGetSoldItemsQuery().Execute(db)
	if err != nil {
		command.PrintErrorf("Error while listing sold items\n")
		return fmt.Errorf("error while listing sold items: %w", err)
	}

	for _, soldItem := range soldItems {
		if err := addToTable(soldItem); err != nil {
			command.PrintErrorf("Error while listing sold items\n")
			return fmt.Errorf("error while listing sold items: %w", err)
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
}
