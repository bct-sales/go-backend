//go:build test

package queries

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestGetSaleItemInformation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, sellCount := range []int{0, 1, 2, 3, 10} {
			label := fmt.Sprintf("Sell count = %d", sellCount)

			t.Run(label, func(t *testing.T) {
				db := OpenInitializedDatabase()
				defer db.Close()

				seller := AddSeller(db)
				cashier := AddCashier(db)
				item := AddItem(db, seller.UserId, 1)

				for i := 0; i < sellCount; i++ {
					AddSaleToDatabase(db, cashier.UserId, []models.Id{item.ItemId})
				}

				itemInformation, err := queries.GetSaleItemInformation(db, item.ItemId)

				if assert.NoError(t, err) {
					assert.Equal(t, item.Description, itemInformation.Description)
					assert.Equal(t, item.PriceInCents, itemInformation.PriceInCents)
					assert.Equal(t, item.CategoryId, itemInformation.ItemCategoryId)
					assert.Equal(t, itemInformation.SellCount, int64(sellCount))
				}
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		nonexistentItemId := models.Id(1)

		itemExists, err := queries.ItemWithIdExists(db, nonexistentItemId)
		if assert.NoError(t, err) {
			if assert.False(t, itemExists) {
				_, err := queries.GetSaleItemInformation(db, 1)
				assert.Error(t, err)
			}
		}
	})
}
