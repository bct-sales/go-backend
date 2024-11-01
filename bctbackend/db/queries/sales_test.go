package queries

import (
	"bctbackend/db/models"
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
