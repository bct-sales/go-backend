//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/test"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestAddSale(t *testing.T) {
	for _, itemIndices := range [][]int{{0}, {1}, {2}, {3}, {0, 1}, {1, 2, 3}, {0, 1, 2, 3}} {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId
		cashierId := test.AddCashierToDatabase(db).UserId

		itemIds := []models.Id{
			test.AddItemToDatabase(db, sellerId, test.WithDummyData(1)).ItemId,
			test.AddItemToDatabase(db, sellerId, test.WithDummyData(2)).ItemId,
			test.AddItemToDatabase(db, sellerId, test.WithDummyData(3)).ItemId,
			test.AddItemToDatabase(db, sellerId, test.WithDummyData(4)).ItemId,
		}

		saleItemIds := make([]models.Id, len(itemIndices))
		for index, itemIndex := range itemIndices {
			saleItemIds[index] = itemIds[itemIndex]
		}

		timestamp := models.NewTimestamp(0)

		saleId, err := queries.AddSale(db, cashierId, timestamp, saleItemIds)
		require.NoError(t, err)

		actualItems, err := queries.GetSaleItems(db, saleId)
		require.NoError(t, err)
		require.Len(t, actualItems, len(saleItemIds))

		for index, actualItem := range actualItems {
			require.Equal(t, saleItemIds[index], actualItem.ItemId)

			expectedItem, err := queries.GetItemWithId(db, saleItemIds[index])
			require.NoError(t, err)
			require.Equal(t, *expectedItem, actualItem)
		}
	}
}

func TestAddSaleWithoutItems(t *testing.T) {
	db := test.OpenInitializedDatabase()
	defer db.Close()

	cashierId := test.AddCashierToDatabase(db).UserId
	timestamp := models.NewTimestamp(0)

	_, err := queries.AddSale(db, cashierId, timestamp, []models.Id{})
	require.Error(t, err)
}

func TestAddSaleWithSellerInsteadOfCashier(t *testing.T) {
	db := test.OpenInitializedDatabase()
	defer db.Close()

	sellerId := test.AddSellerToDatabase(db).UserId
	timestamp := models.NewTimestamp(0)
	itemId := test.AddItemToDatabase(db, sellerId, test.WithDummyData(1)).ItemId

	_, err := queries.AddSale(db, sellerId, timestamp, []models.Id{itemId})
	require.Error(t, err)
}
