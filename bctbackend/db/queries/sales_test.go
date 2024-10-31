package queries

import (
	"bctbackend/db/models"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestAddSale(t *testing.T) {
	db := openInitializedDatabase()

	sellerId := addTestSeller(db)
	cashierId := addTestCashier(db)

	itemIds := []models.Id{
		addTestItem(db, sellerId, 1),
		addTestItem(db, sellerId, 2),
		addTestItem(db, sellerId, 3),
		addTestItem(db, sellerId, 4),
	}

	saleItemIds := []models.Id{
		itemIds[0],
	}

	timestamp := models.NewTimestamp(0)

	saleId, err := AddSale(db, cashierId, timestamp, saleItemIds)

	if assert.NoError(t, err) {
		actualItems, err := GetSaleItems(db, saleId)

		if assert.NoError(t, err) {
			assert.Len(t, actualItems, len(saleItemIds))

			for index, actualItem := range actualItems {
				assert.Equal(t, saleItemIds[index], actualItem.ItemId)

				expectedItem, err := GetItemWithId(db, saleItemIds[index])

				if assert.NoError(t, err) {
					assert.Equal(t, *expectedItem, actualItem)
				}
			}
		}
	}
}
