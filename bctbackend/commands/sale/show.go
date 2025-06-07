package sale

import (
	"bctbackend/commands/common"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type saleShowCommand struct {
	common.Command
}

func NewSaleShowCommand() *cobra.Command {
	var command *saleShowCommand

	command = &saleShowCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "show",
				Short: "Show a sale",
				Long:  `This command shows the details of a specific sale by its ID.`,
				Args:  cobra.ExactArgs(1),
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.Execute(args)
				},
			},
		},
	}

	return command.AsCobraCommand()
}

func (command *saleShowCommand) Execute(args []string) error {
	return command.WithOpenedDatabase(func(db *sql.DB) error {
		saleId, err := command.parseSaleId(args[0])
		if err != nil {
			return err
		}

		sale, err := command.getSaleInformation(db, saleId)
		if err != nil {
			return err
		}

		saleItems, err := command.getSaleItems(db, saleId)
		if err != nil {
			return err
		}

		categoryNameTable, err := command.getCategoryNameTable(db)
		if err != nil {
			return err
		}

		if err := command.printSaleOverview(sale, saleItems); err != nil {
			return err
		}

		if err := command.printSaleItems(saleItems, categoryNameTable); err != nil {
			return err
		}

		return nil
	})
}

func (command *saleShowCommand) parseSaleId(str string) (models.Id, error) {
	saleId, err := models.ParseId(str)

	if err != nil {
		command.PrintErrorf("Invalid sale ID: %v\n", err)
		return 0, fmt.Errorf("invalid sale ID: %w", err)
	}

	return saleId, nil
}

func (command *saleShowCommand) getSaleItems(db *sql.DB, saleId models.Id) ([]models.Item, error) {
	saleItems, err := queries.GetSaleItems(db, saleId)

	if err != nil {
		command.PrintErrorf("An error occurred while getting the sale items: %v\n", err)
		return nil, err
	}

	return saleItems, nil
}

func (command *saleShowCommand) getCategoryNameTable(db *sql.DB) (map[models.Id]string, error) {
	categoryTable, err := queries.GetCategoryNameTable(db)

	if err != nil {
		command.PrintErrorf("An error occurred while fetching categories: %v\n", err)
		return nil, err
	}

	return categoryTable, nil
}

func (command *saleShowCommand) printSaleItems(saleItems []models.Item, categoryNameTable map[models.Id]string) error {
	tableData := pterm.TableData{
		{"ID", "Description", "Price", "Category", "Seller", "Donation", "Charity", "Added At", "Frozen", "Hidden"},
	}

	for _, item := range saleItems {
		categoryName, ok := categoryNameTable[item.CategoryID]
		if !ok {
			command.PrintErrorf("No category found with ID %d for item %d\n", item.CategoryID, item.ItemID)
			return database.ErrNoSuchCategory
		}

		tableData = append(tableData, []string{
			item.ItemID.String(),
			item.Description,
			item.PriceInCents.DecimalNotation(),
			categoryName,
			item.SellerID.String(),
			fmt.Sprintf("%t", item.Donation),
			fmt.Sprintf("%t", item.Charity),
			item.AddedAt.FormattedDateTime(),
			fmt.Sprintf("%t", item.Frozen),
			fmt.Sprintf("%t", item.Hidden),
		})
	}

	if err := pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render(); err != nil {
		command.PrintErrorf("Failed to render table\n")
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}

func (command *saleShowCommand) getSaleInformation(db *sql.DB, saleId models.Id) (*models.Sale, error) {
	sale, err := queries.GetSaleWithId(db, saleId)

	if err != nil {
		if errors.Is(err, database.ErrNoSuchSale) {
			command.PrintErrorf("No sale found with ID %d\n", saleId)
			return nil, err
		}

		command.PrintErrorf("An error occurred while getting the sale information: %v\n", err)
		return nil, err
	}

	return sale, nil
}

func (command *saleShowCommand) printSaleOverview(sale *models.Sale, saleItems []models.Item) error {
	totalCost := models.MoneyInCents(0)
	for _, item := range saleItems {
		totalCost += item.PriceInCents
	}

	tableData := pterm.TableData{
		{"Cashier", sale.CashierID.String()},
		{"Transaction Time", sale.TransactionTime.FormattedDateTime()},
		{"Number of Items", fmt.Sprintf("%d", len(saleItems))},
		{"Total Cost", totalCost.DecimalNotation()},
	}

	if err := pterm.DefaultTable.WithData(tableData).Render(); err != nil {
		command.PrintErrorf("Failed to render sale overview\n")
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}
