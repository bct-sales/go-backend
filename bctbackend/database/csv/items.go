package csv

import (
	models "bctbackend/database/models"
	"encoding/csv"
	"fmt"
	"io"
)

func FormatItemsAsCSV(items []*models.Item, categoryTable map[models.Id]string, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	headers := []string{"item_id", "seller_id", "description", "category", "price_in_cents", "donation", "charity"}
	err := csvWriter.Write(headers)
	if err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	for _, item := range items {
		itemIdString := item.ItemID.String()
		sellerIdString := item.SellerId.String()
		priceString := item.PriceInCents.String()

		categoryString, ok := categoryTable[item.CategoryID]
		if !ok {
			return fmt.Errorf("unknown category id: %v", item.CategoryID)
		}

		var donationString string
		if item.Donation {
			donationString = "true"
		} else {
			donationString = "false"
		}

		var charityString string
		if item.Charity {
			charityString = "true"
		} else {
			charityString = "false"
		}

		err = csvWriter.Write([]string{
			itemIdString,
			sellerIdString,
			item.Description,
			categoryString,
			priceString,
			donationString,
			charityString,
		})

		if err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}
