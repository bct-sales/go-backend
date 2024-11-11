//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestAddSale(t *testing.T) {
	for _, itemIndices := range [][]int{{0}, {1}, {2}, {3}, {0, 1}, {1, 2, 3}, {0, 1, 2, 3}} {
		db := OpenInitializedDatabase()

		sellerId := AddSeller(db).UserId
		cashierId := AddCashier(db).UserId

		itemIds := []models.Id{
			AddItem(db, sellerId, 1).ItemId,
			AddItem(db, sellerId, 2).ItemId,
			AddItem(db, sellerId, 3).ItemId,
			AddItem(db, sellerId, 4).ItemId,
		}

		saleItemIds := make([]models.Id, len(itemIndices))
		for index, itemIndex := range itemIndices {
			saleItemIds[index] = itemIds[itemIndex]
		}

		timestamp := models.NewTimestamp(0)

		saleId, err := queries.AddSale(db, cashierId, timestamp, saleItemIds)

		if assert.NoError(t, err) {
			actualItems, err := queries.GetSaleItems(db, saleId)

			if assert.NoError(t, err) {
				assert.Len(t, actualItems, len(saleItemIds))

				for index, actualItem := range actualItems {
					assert.Equal(t, saleItemIds[index], actualItem.ItemId)

					expectedItem, err := queries.GetItemWithId(db, saleItemIds[index])

					if assert.NoError(t, err) {
						assert.Equal(t, *expectedItem, actualItem)
					}
				}
			}
		}
	}
}

func TestAddSaleWithoutItems(t *testing.T) {
	db := OpenInitializedDatabase()

	cashierId := AddCashier(db).UserId
	timestamp := models.NewTimestamp(0)

	_, err := queries.AddSale(db, cashierId, timestamp, []models.Id{})

	assert.Error(t, err)
}

func TestAddSaleWithSellerInsteadOfCashier(t *testing.T) {
	db := OpenInitializedDatabase()

	sellerId := AddSeller(db).UserId
	timestamp := models.NewTimestamp(0)
	itemId := AddItem(db, sellerId, 1).ItemId

	_, err := queries.AddSale(db, sellerId, timestamp, []models.Id{itemId})

	assert.Error(t, err)
}

func TestGetSales(t *testing.T) {
	db := OpenInitializedDatabase()

	sellerId := AddSeller(db).UserId
	cashierId := AddCashier(db).UserId

	itemIds := []models.Id{
		AddItem(db, sellerId, 1).ItemId,
		AddItem(db, sellerId, 2).ItemId,
		AddItem(db, sellerId, 3).ItemId,
		AddItem(db, sellerId, 4).ItemId,
	}

	saleIds := make([]models.Id, len(itemIds))
	for _, itemId := range itemIds {
		AddSaleToDatabase(db, cashierId, []models.Id{itemId})
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
	db := OpenInitializedDatabase()

	sellerId := AddSeller(db).UserId
	cashierId := AddCashier(db).UserId
	itemId := AddItem(db, sellerId, 1).ItemId

	saleId := AddSaleToDatabase(db, cashierId, []models.Id{itemId})
	saleExists, err := queries.SaleExists(db, saleId)

	if assert.NoError(t, err) {
		assert.True(t, saleExists)
	}
}

func TestGetSaleItems(t *testing.T) {
	db := OpenInitializedDatabase()

	sellerId := AddSeller(db).UserId
	cashierId := AddCashier(db).UserId
	itemIds := []models.Id{
		AddItem(db, sellerId, 1).ItemId,
		AddItem(db, sellerId, 2).ItemId,
		AddItem(db, sellerId, 3).ItemId,
		AddItem(db, sellerId, 4).ItemId,
	}

	saleId := AddSaleToDatabase(db, cashierId, itemIds)

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
	db := OpenInitializedDatabase()

	sellerId := AddSeller(db).UserId
	cashierId := AddCashier(db).UserId
	sale1ItemIds := []models.Id{
		AddItem(db, sellerId, 1).ItemId,
		AddItem(db, sellerId, 2).ItemId,
	}
	sale2ItemIds := []models.Id{
		AddItem(db, sellerId, 3).ItemId,
		AddItem(db, sellerId, 4).ItemId,
	}

	sale1Id := AddSaleToDatabase(db, cashierId, sale1ItemIds)
	sale2Id := AddSaleToDatabase(db, cashierId, sale2ItemIds)

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
	db := OpenInitializedDatabase()

	err := queries.RemoveSale(db, 0)

	assert.ErrorIs(t, err, queries.NoSuchSaleError{})
}
