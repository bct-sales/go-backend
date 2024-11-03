package queries

import (
	"bctbackend/database/models"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestAddSale(t *testing.T) {
	for _, itemIndices := range [][]int{{0}, {1}, {2}, {3}, {0, 1}, {1, 2, 3}, {0, 1, 2, 3}} {
		db := openInitializedDatabase()

		sellerId := addTestSeller(db)
		cashierId := addTestCashier(db)

		itemIds := []models.Id{
			addTestItem(db, sellerId, 1),
			addTestItem(db, sellerId, 2),
			addTestItem(db, sellerId, 3),
			addTestItem(db, sellerId, 4),
		}

		saleItemIds := make([]models.Id, len(itemIndices))
		for index, itemIndex := range itemIndices {
			saleItemIds[index] = itemIds[itemIndex]
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
}

func TestAddSaleWithoutItems(t *testing.T) {
	db := openInitializedDatabase()

	cashierId := addTestCashier(db)
	timestamp := models.NewTimestamp(0)

	_, err := AddSale(db, cashierId, timestamp, []models.Id{})

	assert.Error(t, err)
}

func TestAddSaleWithSellerInsteadOfCashier(t *testing.T) {
	db := openInitializedDatabase()

	sellerId := addTestSeller(db)
	timestamp := models.NewTimestamp(0)
	itemId := addTestItem(db, sellerId, 1)

	_, err := AddSale(db, sellerId, timestamp, []models.Id{itemId})

	assert.Error(t, err)
}

func TestGetSales(t *testing.T) {
	db := openInitializedDatabase()

	sellerId := addTestSeller(db)
	cashierId := addTestCashier(db)

	itemIds := []models.Id{
		addTestItem(db, sellerId, 1),
		addTestItem(db, sellerId, 2),
		addTestItem(db, sellerId, 3),
		addTestItem(db, sellerId, 4),
	}

	saleIds := make([]models.Id, len(itemIds))
	for _, itemId := range itemIds {
		addTestSale(db, cashierId, []models.Id{itemId})
	}

	actualSales, err := GetSales(db)

	if assert.NoError(t, err) {
		assert.Len(t, actualSales, len(saleIds))

		for _, actualSale := range actualSales {
			assert.Equal(t, cashierId, actualSale.CashierId)

			saleItems, err := GetSaleItems(db, actualSale.SaleId)

			if assert.NoError(t, err) {
				assert.Equal(t, 1, len(saleItems))
			}
		}
	}
}

func TestSaleExists(t *testing.T) {
	db := openInitializedDatabase()

	sellerId := addTestSeller(db)
	cashierId := addTestCashier(db)
	itemId := addTestItem(db, sellerId, 1)

	saleId := addTestSale(db, cashierId, []models.Id{itemId})
	saleExists, err := SaleExists(db, saleId)

	if assert.NoError(t, err) {
		assert.True(t, saleExists)
	}
}

func TestGetSaleItems(t *testing.T) {
	db := openInitializedDatabase()

	sellerId := addTestSeller(db)
	cashierId := addTestCashier(db)
	itemIds := []models.Id{
		addTestItem(db, sellerId, 1),
		addTestItem(db, sellerId, 2),
		addTestItem(db, sellerId, 3),
		addTestItem(db, sellerId, 4),
	}

	saleId := addTestSale(db, cashierId, itemIds)

	actualItems, err := GetSaleItems(db, saleId)

	if assert.NoError(t, err) {
		assert.Len(t, actualItems, len(itemIds))

		for index, actualItem := range actualItems {
			assert.Equal(t, itemIds[index], actualItem.ItemId)

			expectedItem, err := GetItemWithId(db, itemIds[index])

			if assert.NoError(t, err) {
				assert.Equal(t, *expectedItem, actualItem)
			}
		}
	}
}

func TestRemoveSale(t *testing.T) {
	db := openInitializedDatabase()

	sellerId := addTestSeller(db)
	cashierId := addTestCashier(db)
	sale1ItemIds := []models.Id{
		addTestItem(db, sellerId, 1),
		addTestItem(db, sellerId, 2),
	}
	sale2ItemIds := []models.Id{
		addTestItem(db, sellerId, 3),
		addTestItem(db, sellerId, 4),
	}

	sale1Id := addTestSale(db, cashierId, sale1ItemIds)
	sale2Id := addTestSale(db, cashierId, sale2ItemIds)

	err := RemoveSale(db, sale1Id)

	if assert.NoError(t, err) {
		sale1Exists, err := SaleExists(db, sale1Id)

		if assert.NoError(t, err) {
			assert.False(t, sale1Exists)
		}

		sale2Exists, err := SaleExists(db, sale2Id)

		if assert.NoError(t, err) {
			assert.True(t, sale2Exists)
		}
	}
}

func TestRemoveNonexistentSale(t *testing.T) {
	db := openInitializedDatabase()

	err := RemoveSale(db, 0)

	assert.ErrorIs(t, err, NoSuchSaleError{})
}
