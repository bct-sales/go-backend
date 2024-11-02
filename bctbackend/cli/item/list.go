package user

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"fmt"
	"strconv"

	"database/sql"

	"github.com/pterm/pterm"
	_ "modernc.org/sqlite"
)

func ListItems(databasePath string) error {
	db, err := sql.Open("sqlite", databasePath)

	if err != nil {
		return fmt.Errorf("error while opening database: %v", err)
	}

	defer db.Close()

	items, err := queries.GetItems(db)

	if err != nil {
		return fmt.Errorf("error while listing items: %v", err)
	}

	tableData := pterm.TableData{
		{"ID", "Description", "Price", "Category", "Seller", "Donation", "Charity"},
	}

	itemCount := 0

	for _, item := range items {
		itemIdString := strconv.FormatInt(item.ItemId, 10)
		itemDescriptionString := item.Description
		itemPriceString := strconv.FormatFloat(float64(item.PriceInCents)/100.0, 'f', 2, 64)
		itemSellerIdString := strconv.FormatInt(item.SellerId, 10)
		itemDonationString := strconv.FormatBool(item.Donation)
		itemCharityString := strconv.FormatBool(item.Charity)
		itemCategoryNameString, err := models.NameOfCategory(item.CategoryId)

		if err != nil {
			return fmt.Errorf("error while converting role to string: %v", err)
		}

		tableData = append(tableData, []string{
			itemIdString,
			itemDescriptionString,
			itemPriceString,
			itemCategoryNameString,
			itemSellerIdString,
			itemDonationString,
			itemCharityString,
		})

		itemCount++
	}

	if itemCount == 0 {
		fmt.Println("No items found")

		return nil
	} else {
		err = pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

		if err != nil {
			return fmt.Errorf("error while rendering table: %v", err)
		}

		fmt.Printf("Number of items listed: %d\n", itemCount)

		return nil
	}
}
