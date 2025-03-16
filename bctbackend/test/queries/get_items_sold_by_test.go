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

func TestGetItemsSoldBy(t *testing.T) {
	t.Run("Zero items sold", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId
		cashierId := AddCashierToDatabase(db).UserId
		AddItemToDatabase(db, sellerId, WithDummyData(1))
		AddItemToDatabase(db, sellerId, WithDummyData(2))
		AddItemToDatabase(db, sellerId, WithDummyData(3))
		AddItemToDatabase(db, sellerId, WithDummyData(4))

		items, err := queries.GetItemsSoldBy(db, cashierId)
		require.NoError(t, err)
		require.Len(t, items, 0)
	})

	t.Run("Items sold by different cashier", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId
		zeroSaleCashier := AddCashierToDatabase(db).UserId
		cashierWithSalesId := AddCashierToDatabase(db).UserId

		itemIds := []models.Id{
			AddItemToDatabase(db, sellerId, WithDummyData(1)).ItemId,
			AddItemToDatabase(db, sellerId, WithDummyData(2)).ItemId,
			AddItemToDatabase(db, sellerId, WithDummyData(3)).ItemId,
			AddItemToDatabase(db, sellerId, WithDummyData(4)).ItemId,
		}

		AddSaleToDatabase(db, cashierWithSalesId, itemIds)

		items, err := queries.GetItemsSoldBy(db, zeroSaleCashier)
		require.NoError(t, err)
		require.Len(t, items, 0)
	})

	t.Run("Four items in single sale", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId
		cashier := AddCashierToDatabase(db).UserId

		expectedItems := []models.Item{
			*AddItemToDatabase(db, sellerId, WithDummyData(1)),
			*AddItemToDatabase(db, sellerId, WithDummyData(2)),
			*AddItemToDatabase(db, sellerId, WithDummyData(3)),
			*AddItemToDatabase(db, sellerId, WithDummyData(4)),
		}

		AddSaleToDatabase(db, cashier, []models.Id{expectedItems[0].ItemId, expectedItems[1].ItemId, expectedItems[2].ItemId, expectedItems[3].ItemId})

		actualItems, err := queries.GetItemsSoldBy(db, cashier)
		require.NoError(t, err)
		require.Len(t, actualItems, 4)
		require.Equal(t, expectedItems, actualItems)
	})

	t.Run("Four items in single sale, reordered", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId
		cashier := AddCashierToDatabase(db).UserId

		item1 := *AddItemToDatabase(db, sellerId, WithDummyData(1))
		item2 := *AddItemToDatabase(db, sellerId, WithDummyData(2))
		item3 := *AddItemToDatabase(db, sellerId, WithDummyData(3))
		item4 := *AddItemToDatabase(db, sellerId, WithDummyData(4))

		expectedItems := []models.Item{item1, item2, item3, item4}
		AddSaleToDatabase(db, cashier, []models.Id{item4.ItemId, item3.ItemId, item2.ItemId, item1.ItemId})

		actualItems, err := queries.GetItemsSoldBy(db, cashier)
		require.NoError(t, err)
		require.Len(t, expectedItems, 4)
		require.Equal(t, expectedItems, actualItems)
	})

	t.Run("Four items in separate sales", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId
		cashier := AddCashierToDatabase(db).UserId

		item1 := *AddItemToDatabase(db, sellerId, WithDummyData(1))
		item2 := *AddItemToDatabase(db, sellerId, WithDummyData(2))
		item3 := *AddItemToDatabase(db, sellerId, WithDummyData(3))
		item4 := *AddItemToDatabase(db, sellerId, WithDummyData(4))

		AddSaleToDatabase(db, cashier, []models.Id{item1.ItemId, item2.ItemId})
		AddSaleToDatabase(db, cashier, []models.Id{item3.ItemId, item4.ItemId})

		items, err := queries.GetItemsSoldBy(db, cashier)
		require.NoError(t, err)
		require.Len(t, items, 4)
		require.Equal(t, []models.Item{item1, item2, item3, item4}, items)
	})

	t.Run("Four items in separate sales, reordered", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId
		cashier := AddCashierToDatabase(db).UserId

		item1 := *AddItemToDatabase(db, sellerId, WithDummyData(1))
		item2 := *AddItemToDatabase(db, sellerId, WithDummyData(2))
		item3 := *AddItemToDatabase(db, sellerId, WithDummyData(3))
		item4 := *AddItemToDatabase(db, sellerId, WithDummyData(4))

		AddSaleToDatabase(db, cashier, []models.Id{item2.ItemId, item1.ItemId}, WithTransactionTime(models.NewTimestamp(1)))
		AddSaleToDatabase(db, cashier, []models.Id{item4.ItemId, item3.ItemId}, WithTransactionTime(models.NewTimestamp(0)))

		items, err := queries.GetItemsSoldBy(db, cashier)
		require.NoError(t, err)
		require.Len(t, items, 4)
		require.Equal(t, []models.Item{item1, item2, item3, item4}, items)
	})

	t.Run("Four items in separate sales, reordered 2", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId
		cashier := AddCashierToDatabase(db).UserId

		item1 := *AddItemToDatabase(db, sellerId, WithDummyData(1))
		item2 := *AddItemToDatabase(db, sellerId, WithDummyData(2))
		item3 := *AddItemToDatabase(db, sellerId, WithDummyData(3))
		item4 := *AddItemToDatabase(db, sellerId, WithDummyData(4))

		AddSaleToDatabase(db, cashier, []models.Id{item2.ItemId, item1.ItemId}, WithTransactionTime(models.NewTimestamp(0)))
		AddSaleToDatabase(db, cashier, []models.Id{item4.ItemId, item3.ItemId}, WithTransactionTime(models.NewTimestamp(1)))

		items, err := queries.GetItemsSoldBy(db, cashier)
		require.NoError(t, err)
		require.Len(t, items, 4)
		require.Equal(t, []models.Item{item3, item4, item1, item2}, items)
	})

	t.Run("Cashier does not exist", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		cashier := AddCashierToDatabase(db).UserId
		unknownCashierId := cashier + 1

		{
			userExists, err := queries.UserWithIdExists(db, unknownCashierId)
			require.NoError(t, err)
			require.False(t, userExists)
		}

		_, err := queries.GetItemsSoldBy(db, unknownCashierId)
		var noSuchUserError *queries.UnknownUserError
		require.ErrorAs(t, err, &noSuchUserError)
	})

	t.Run("User has wrong role", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId

		_, err := queries.GetItemsSoldBy(db, sellerId)
		var invalidRoleError *queries.InvalidRoleError
		require.ErrorAs(t, err, &invalidRoleError)
	})
}
