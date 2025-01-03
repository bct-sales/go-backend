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

func TestAddSale(t *testing.T) {
	for _, itemIndices := range [][]int{{0}, {1}, {2}, {3}, {0, 1}, {1, 2, 3}, {0, 1, 2, 3}} {
		db := test.OpenInitializedDatabase()

		sellerId := test.AddSellerToDatabase(db).UserId
		cashierId := test.AddCashierToDatabase(db).UserId

		itemIds := []models.Id{
			test.AddItemToDatabase(db, sellerId, 1).ItemId,
			test.AddItemToDatabase(db, sellerId, 2).ItemId,
			test.AddItemToDatabase(db, sellerId, 3).ItemId,
			test.AddItemToDatabase(db, sellerId, 4).ItemId,
		}

		saleItemIds := make([]models.Id, len(itemIndices))
		for index, itemIndex := range itemIndices {
			saleItemIds[index] = itemIds[itemIndex]
		}

		timestamp := models.NewTimestamp(0)

		saleId, err := queries.AddSale(db, cashierId, timestamp, saleItemIds)

		if !assert.NoError(t, err) {
			return
		}

		actualItems, err := queries.GetSaleItems(db, saleId)

		if !assert.NoError(t, err) {
			return
		}

		if !assert.Len(t, actualItems, len(saleItemIds)) {
			return
		}

		for index, actualItem := range actualItems {
			if !assert.Equal(t, saleItemIds[index], actualItem.ItemId) {
				return
			}

			expectedItem, err := queries.GetItemWithId(db, saleItemIds[index])

			if !assert.NoError(t, err) {
				return
			}

			if !assert.Equal(t, *expectedItem, actualItem) {
				return
			}
		}
	}
}

func TestAddSaleWithoutItems(t *testing.T) {
	db := test.OpenInitializedDatabase()

	cashierId := test.AddCashierToDatabase(db).UserId
	timestamp := models.NewTimestamp(0)

	_, err := queries.AddSale(db, cashierId, timestamp, []models.Id{})

	if !assert.Error(t, err) {
		return
	}
}

func TestAddSaleWithSellerInsteadOfCashier(t *testing.T) {
	db := test.OpenInitializedDatabase()

	sellerId := test.AddSellerToDatabase(db).UserId
	timestamp := models.NewTimestamp(0)
	itemId := test.AddItemToDatabase(db, sellerId, 1).ItemId

	_, err := queries.AddSale(db, sellerId, timestamp, []models.Id{itemId})

	if !assert.Error(t, err) {
		return
	}
}
