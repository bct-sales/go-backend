package csv

import (
	models "bctbackend/database/models"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

func FormatItemsAsCSV(items []*models.Item, categoryNameTable map[models.ID]string, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	headers := []string{"item_id", "seller_id", "description", "category", "price_in_cents", "donation", "charity", "large"}
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

		err = csvWriter.Write([]string{
			itemIDString,
			sellerIDString,
			item.Description,
			categoryString,
			priceString,
			donationString,
			charityString,
			largeString,
		})

		if err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}
