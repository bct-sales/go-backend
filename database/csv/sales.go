package csv

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

func FormatSoldItemsAsCSV(soldItems []*queries.SoldItem, categoryNameTable map[models.ID]string, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	headers := []string{"sale_id", "cashier_id", "transaction_time", "item_id", "description", "price_in_cents", "item_category", "seller_id", "donation", "charity"}
	err := csvWriter.Write(headers)
	if err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	for _, soldItem := range soldItems {
		saleIdStr := soldItem.SaleId.String()
		cashierIdStr := soldItem.CashierId.String()
		transactionTimeStr := soldItem.TransactionTime.FormattedDateTime()
		itemIdStr := soldItem.ItemID.String()
		descriptionStr := soldItem.Description
		priceInCentsStr := soldItem.PriceInCents.String()
		itemCategoryStr, ok := categoryNameTable[soldItem.ItemCategoryId]
		if !ok {
			return fmt.Errorf("unknown category id: %v", soldItem.ItemCategoryId)
		}
		sellerIdStr := soldItem.SellerId.String()
		donationStr := strconv.FormatBool(soldItem.Donation)
		charityStr := strconv.FormatBool(soldItem.Charity)

		err = csvWriter.Write([]string{
			saleIdStr,
			cashierIdStr,
			transactionTimeStr,
			itemIdStr,
			descriptionStr,
			priceInCentsStr,
			itemCategoryStr,
			sellerIdStr,
			donationStr,
			charityStr,
		})

		if err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}
