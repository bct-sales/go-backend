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
		saleIDStr := soldItem.SaleID.String()
		cashierIDStr := soldItem.CashierID.String()
		transactionTimeStr := soldItem.TransactionTime.FormattedDateTime()
		itemIDStr := soldItem.ItemID.String()
		descriptionStr := soldItem.Description
		priceInCentsStr := soldItem.PriceInCents.String()
		itemCategoryStr, ok := categoryNameTable[soldItem.ItemCategoryID]
		if !ok {
			return fmt.Errorf("unknown category id: %v", soldItem.ItemCategoryID)
		}
		sellerIDStr := soldItem.SellerID.String()
		donationStr := strconv.FormatBool(soldItem.Donation)
		charityStr := strconv.FormatBool(soldItem.Charity)

		err = csvWriter.Write([]string{
			saleIDStr,
			cashierIDStr,
			transactionTimeStr,
			itemIDStr,
			descriptionStr,
			priceInCentsStr,
			itemCategoryStr,
			sellerIDStr,
			donationStr,
			charityStr,
		})

		if err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}
