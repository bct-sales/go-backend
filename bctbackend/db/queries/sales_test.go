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

	itemIds := []models.Id{addTestItem(db, sellerId, 1), addTestItem(db, sellerId, 2)}
	timestamp := models.NewTimestamp(0)

	saleId, err := AddSale(db, cashierId, timestamp, itemIds)

	if assert.NoError(t, err) {
		items, err := GetSaleItems(db, saleId)

		if assert.NoError(t, err) {
			assert.Len(t, items, 2)
			assert.Equal(t, itemIds[0], items[0].ItemId)
			assert.Equal(t, itemIds[1], items[1].ItemId)
		}
	}
}
