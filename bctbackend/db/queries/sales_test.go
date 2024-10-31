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
		items, err := GetSaleItems(db, saleId)

		if assert.NoError(t, err) {
			assert.Len(t, items, len(saleItemIds))

			for index, item := range items {
				assert.Equal(t, saleItemIds[index], item.ItemId)
			}
		}
	}
}
