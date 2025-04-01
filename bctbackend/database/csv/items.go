package csv

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/defs"
	"database/sql"
	"encoding/csv"
	"io"
)

func OutputItems(db *sql.DB, writer io.Writer) (r_err error) {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	headers := []string{"item_id", "seller_id", "description", "category", "price_in_cents", "donation", "charity"}
	err := csvWriter.Write(headers)
	if err != nil {
		return err
	}

	items := []*models.Item{}
	if err := queries.GetItems(db, queries.CollectTo(&items)); err != nil {
		return err
	}

	for _, item := range items {
		itemIdString := models.IdToString(item.ItemId)
		sellerIdString := models.IdToString(item.SellerId)
		priceString := models.MoneyInCentsToString(item.PriceInCents)

		categoryString, err := defs.NameOfCategory(item.CategoryId)
		if err != nil {
			return err
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
			return err
		}
	}

	return nil
}
