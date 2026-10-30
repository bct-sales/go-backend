package item

import (
	"bctbackend/commands/common"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type listItemsCommand struct {
	common.Command
	showHidden bool   `exhaustruct:"optional"`
	format     string `exhaustruct:"optional"`
}

func NewListItemsCommand() *cobra.Command {
	var command *listItemsCommand

	command = &listItemsCommand{
		Command: common.Command{
			CobraCommand: &cobra.Command{
				Use:   "list",
				Short: "List all items",
				Long:  `This command lists all items in the database.`,
				RunE: func(cmd *cobra.Command, args []string) error {
					return command.execute(cmd.Context())
				},
				Args: cobra.NoArgs,
			},
		},
	}

	command.CobraCommand.Flags().BoolVar(&command.showHidden, "show-hidden", false, "Show hidden items")
	command.CobraCommand.Flags().StringVar(&command.format, "format", "table", "Output format (format, csv)")

	return command.AsCobraCommand()
}

func (c *listItemsCommand) execute(ctx context.Context) error {
	switch c.format {
	case "table":
		return c.listItemsInTableFormat(ctx)
	case "csv":
		return c.listItemsInCSVFormat(ctx)
	default:
		c.PrintErrorf("Invalid format: %s\n", c.format)
		return fmt.Errorf("unknown format: %s", c.format)
	}
}

func (c *listItemsCommand) createQuery() *queries.ListItemsQuery {
	query := queries.ListItems()

	if !c.showHidden {
		query.WithHidden(false)
	}

	return query
}

func (c *listItemsCommand) listItemsInTableFormat(ctx context.Context) error {
	return c.WithOpenedDatabase(ctx, func(_ context.Context, db *sql.DB) error {
		query := c.createQuery()

		items := []*queries.ListItemsQueryRow{}
		if err := query.Execute(db, queries.CollectTo(&items)); err != nil {
			c.PrintErrorf("Error while getting items: %v\n", err)
			return err
		}

		itemCount := len(items)

		if itemCount > 0 {
			categoryNameTable, err := c.GetCategoryNameTable(db)
			if err != nil {
				return err
			}

			tableData := pterm.TableData{
				{"ID", "Description", "Price", "Category", "Seller", "Donation", "Charity", "Large", "Added At", "Frozen", "Hidden", "SaleCount"},
			}

			for _, item := range items {
				categoryName, ok := categoryNameTable[item.CategoryID]
				if !ok {
					return dberr.ErrNoSuchCategory
				}

				tableData = append(tableData, []string{
					item.ItemID.String(),
					item.Description,
					item.PriceInCents.DecimalNotation(),
					categoryName,
					item.SellerID.String(),
					strconv.FormatBool(item.Donation),
					strconv.FormatBool(item.Charity),
					strconv.FormatBool(item.Large),
					item.AddedAt.FormattedDateTime(),
					strconv.FormatBool(item.Frozen),
					strconv.FormatBool(item.Hidden),
					strconv.FormatInt(int64(item.SaleCount), 10),
				})
			}

			if err := pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(tableData).Render(); err != nil {
				c.PrintErrorf("Error while rendering table\n")
				return fmt.Errorf("failed to render table: %w", err)
			}

			c.Printf("Number of items listed: %d\n", itemCount)
			return nil
		} else {
			c.Printf("No items found")
			return nil
		}
	})
}

func (c *listItemsCommand) listItemsInCSVFormat(ctx context.Context) error {
	return c.WithOpenedDatabase(ctx, func(_ context.Context, db *sql.DB) error {
		query := c.createQuery()

		items := []*queries.ListItemsQueryRow{}
		if err := query.Execute(db, queries.CollectTo(&items)); err != nil {
			return fmt.Errorf("failed to get items: %w", err)
		}

		categoryNameTable, err := c.GetCategoryNameTable(db)
		if err != nil {
			return err
		}

		err = c.writeCSVFile(items, categoryNameTable, os.Stdout)
		if err != nil {
			c.PrintErrorf("Error while formatting items as CSV\n")
			return fmt.Errorf("failed to format items as a CSV: %w", err)
		}

		return nil
	})
}

func (c *listItemsCommand) writeCSVFile(items []*queries.ListItemsQueryRow, categoryNameTable map[models.ID]string, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	headers := []string{"item_id", "seller_id", "description", "category", "price_in_cents", "donation", "charity", "large", "sale_count"}
	err := csvWriter.Write(headers)
	if err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	for _, item := range items {
		itemIDString := item.ItemID.String()
		sellerIDString := item.SellerID.String()
		priceString := item.PriceInCents.String()

		categoryString, ok := categoryNameTable[item.CategoryID]
		if !ok {
			return fmt.Errorf("unknown category id: %v", item.CategoryID)
		}

		donationString := strconv.FormatBool(item.Donation)
		charityString := strconv.FormatBool(item.Charity)
		largeString := strconv.FormatBool(item.Large)
		saleCount := strconv.FormatInt(int64(item.SaleCount), 10)

		err = csvWriter.Write([]string{
			itemIDString,
			sellerIDString,
			item.Description,
			categoryString,
			priceString,
			donationString,
			charityString,
			largeString,
			saleCount,
		})

		if err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}
