//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/test"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestGetSaleItems(t *testing.T) {
	db := test.OpenInitializedDatabase()

	sellerId := test.AddSellerToDatabase(db).UserId
	cashierId := test.AddCashierToDatabase(db).UserId
	itemIds := []models.Id{
		test.AddItemToDatabase(db, sellerId, 1).ItemId,
		test.AddItemToDatabase(db, sellerId, 2).ItemId,
		test.AddItemToDatabase(db, sellerId, 3).ItemId,
		test.AddItemToDatabase(db, sellerId, 4).ItemId,
	}

	saleId := test.AddSaleToDatabase(db, cashierId, itemIds)

	actualItems, err := queries.GetSaleItems(db, saleId)

	if assert.NoError(t, err) {
		assert.Len(t, actualItems, len(itemIds))

		for index, actualItem := range actualItems {
			assert.Equal(t, itemIds[index], actualItem.ItemId)

			expectedItem, err := queries.GetItemWithId(db, itemIds[index])

			if assert.NoError(t, err) {
				assert.Equal(t, *expectedItem, actualItem)
			}
		}
	}
}
