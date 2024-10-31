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

	item1 := addTestItem(db, sellerId, 1)
	item2 := addTestItem(db, sellerId, 2)

	timestamp := models.NewTimestamp(0)
	itemIds := []models.Id{item1, item2}

	saleId, err := AddSale(db, cashierId, timestamp, itemIds)

	if assert.NoError(t, err) {
		items, err := GetSaleItems(db, saleId)

		if assert.NoError(t, err) {
			assert.Len(t, items, 2)
			assert.Equal(t, item1, items[0].ItemId)
			assert.Equal(t, item2, items[1].ItemId)
		}
	}
}
