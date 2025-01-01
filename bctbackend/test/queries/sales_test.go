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

func TestGetSales(t *testing.T) {
	db := test.OpenInitializedDatabase()

	sellerId := test.AddSellerToDatabase(db).UserId
	cashierId := test.AddCashierToDatabase(db).UserId

	itemIds := []models.Id{
		test.AddItemToDatabase(db, sellerId, 1).ItemId,
		test.AddItemToDatabase(db, sellerId, 2).ItemId,
		test.AddItemToDatabase(db, sellerId, 3).ItemId,
		test.AddItemToDatabase(db, sellerId, 4).ItemId,
	}

	saleIds := make([]models.Id, len(itemIds))
	for _, itemId := range itemIds {
		test.AddSaleToDatabase(db, cashierId, []models.Id{itemId})
	}

	actualSales, err := queries.GetSales(db)

	if assert.NoError(t, err) {
		assert.Len(t, actualSales, len(saleIds))

		for _, actualSale := range actualSales {
			assert.Equal(t, cashierId, actualSale.CashierId)

			saleItems, err := queries.GetSaleItems(db, actualSale.SaleId)

			if assert.NoError(t, err) {
				assert.Equal(t, 1, len(saleItems))
			}
		}
	}
}

func TestSaleExists(t *testing.T) {
	db := test.OpenInitializedDatabase()

	sellerId := test.AddSellerToDatabase(db).UserId
	cashierId := test.AddCashierToDatabase(db).UserId
	itemId := test.AddItemToDatabase(db, sellerId, 1).ItemId

	saleId := test.AddSaleToDatabase(db, cashierId, []models.Id{itemId})
	saleExists, err := queries.SaleExists(db, saleId)

	if assert.NoError(t, err) {
		assert.True(t, saleExists)
	}
}

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

func TestRemoveSale(t *testing.T) {
	db := test.OpenInitializedDatabase()

	sellerId := test.AddSellerToDatabase(db).UserId
	cashierId := test.AddCashierToDatabase(db).UserId
	sale1ItemIds := []models.Id{
		test.AddItemToDatabase(db, sellerId, 1).ItemId,
		test.AddItemToDatabase(db, sellerId, 2).ItemId,
	}
	sale2ItemIds := []models.Id{
		test.AddItemToDatabase(db, sellerId, 3).ItemId,
		test.AddItemToDatabase(db, sellerId, 4).ItemId,
	}

	sale1Id := test.AddSaleToDatabase(db, cashierId, sale1ItemIds)
	sale2Id := test.AddSaleToDatabase(db, cashierId, sale2ItemIds)

	err := queries.RemoveSale(db, sale1Id)

	if assert.NoError(t, err) {
		sale1Exists, err := queries.SaleExists(db, sale1Id)

		if assert.NoError(t, err) {
			assert.False(t, sale1Exists)
		}

		sale2Exists, err := queries.SaleExists(db, sale2Id)

		if assert.NoError(t, err) {
			assert.True(t, sale2Exists)
		}
	}
}

func TestRemoveNonexistentSale(t *testing.T) {
	db := test.OpenInitializedDatabase()

	err := queries.RemoveSale(db, 0)

	assert.ErrorIs(t, err, queries.NoSuchSaleError{})
}
