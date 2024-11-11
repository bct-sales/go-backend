//go:build test

package queries

import (
	models "bctbackend/database/models"
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
				db := openInitializedDatabase()
				defer db.Close()

				seller := addTestSeller(db)
				cashier := addTestCashier(db)
				item := addTestItem(db, seller.UserId, 1)

				for i := 0; i < sellCount; i++ {
					addTestSale(db, cashier.UserId, []models.Id{item.ItemId})
				}

				itemInformation, err := GetSaleItemInformation(db, item.ItemId)

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
		db := openInitializedDatabase()
		defer db.Close()

		nonexistentItemId := models.Id(1)

		itemExists, err := ItemWithIdExists(db, nonexistentItemId)
		if assert.NoError(t, err) {
			if assert.False(t, itemExists) {
				_, err := GetSaleItemInformation(db, 1)
				assert.Error(t, err)
			}
		}
	})
}
