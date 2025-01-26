//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestAddSale(t *testing.T) {
	for _, itemIndices := range [][]int{{0}, {1}, {2}, {3}, {0, 1}, {1, 2, 3}, {0, 1, 2, 3}} {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId
		cashierId := AddCashierToDatabase(db).UserId

		itemIds := []models.Id{
			AddItemToDatabase(db, sellerId, WithDummyData(1)).ItemId,
			AddItemToDatabase(db, sellerId, WithDummyData(2)).ItemId,
			AddItemToDatabase(db, sellerId, WithDummyData(3)).ItemId,
			AddItemToDatabase(db, sellerId, WithDummyData(4)).ItemId,
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
	db := OpenInitializedDatabase()
	defer db.Close()

	cashierId := AddCashierToDatabase(db).UserId
	timestamp := models.NewTimestamp(0)

	_, err := queries.AddSale(db, cashierId, timestamp, []models.Id{})
	require.Error(t, err)
}

func TestAddSaleWithSellerInsteadOfCashier(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	sellerId := AddSellerToDatabase(db).UserId
	timestamp := models.NewTimestamp(0)
	itemId := AddItemToDatabase(db, sellerId, WithDummyData(1)).ItemId

	_, err := queries.AddSale(db, sellerId, timestamp, []models.Id{itemId})
	require.Error(t, err)
}

func TestAddSaleWithSameItemTwice(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	sellerId := AddSellerToDatabase(db).UserId
	timestamp := models.NewTimestamp(0)
	item := AddItemToDatabase(db, sellerId, WithDummyData(1))

	_, err := queries.AddSale(db, sellerId, timestamp, []models.Id{item.ItemId, item.ItemId})
	require.Error(t, err)
}
