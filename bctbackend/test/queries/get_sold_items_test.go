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

func TestGetSoldItems(t *testing.T) {
	t.Run("No items in existence", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		soldItems, err := queries.GetSoldItems(db)
		require.NoError(t, err)
		require.Empty(t, soldItems)
	})

	t.Run("Single unsold item", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		seller := AddSellerToDatabase(db)
		AddItemToDatabase(db, seller.UserId, WithDummyData(1))

		soldItems, err := queries.GetSoldItems(db)
		require.NoError(t, err)
		require.Empty(t, soldItems)
	})

	t.Run("Single sold item", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		seller := AddSellerToDatabase(db)
		cashier := AddCashierToDatabase(db)
		item := AddItemToDatabase(db, seller.UserId, WithDummyData(1))
		AddSaleToDatabase(db, cashier.UserId, []models.Id{item.ItemId})

		soldItems, err := queries.GetSoldItems(db)
		require.NoError(t, err)
		require.Len(t, soldItems, 1)
		require.Equal(t, item, soldItems[0])
	})

	t.Run("Doubly sold item", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		seller := AddSellerToDatabase(db)
		cashier := AddCashierToDatabase(db)
		item := AddItemToDatabase(db, seller.UserId, WithDummyData(1))
		AddSaleToDatabase(db, cashier.UserId, []models.Id{item.ItemId})
		AddSaleToDatabase(db, cashier.UserId, []models.Id{item.ItemId})

		soldItems, err := queries.GetSoldItems(db)
		require.NoError(t, err)
		require.Len(t, soldItems, 1)
		require.Equal(t, item, soldItems[0])
	})

	t.Run("Two sold items in single sale", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		seller := AddSellerToDatabase(db)
		cashier := AddCashierToDatabase(db)
		item1 := AddItemToDatabase(db, seller.UserId, WithDummyData(1))
		item2 := AddItemToDatabase(db, seller.UserId, WithDummyData(2))
		AddSaleToDatabase(db, cashier.UserId, []models.Id{item1.ItemId, item2.ItemId})

		soldItems, err := queries.GetSoldItems(db)
		require.NoError(t, err)
		require.Len(t, soldItems, 2)
		require.Equal(t, item1, soldItems[0])
		require.Equal(t, item2, soldItems[1])
	})

	t.Run("Two sold items in separate sales", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		seller := AddSellerToDatabase(db)
		cashier := AddCashierToDatabase(db)
		item1 := AddItemToDatabase(db, seller.UserId, WithDummyData(1))
		item2 := AddItemToDatabase(db, seller.UserId, WithDummyData(2))
		AddSaleToDatabase(db, cashier.UserId, []models.Id{item1.ItemId}, WithTransactionTime(models.Timestamp(100)))
		AddSaleToDatabase(db, cashier.UserId, []models.Id{item2.ItemId}, WithTransactionTime(models.Timestamp(200)))

		soldItems, err := queries.GetSoldItems(db)
		require.NoError(t, err)
		require.Len(t, soldItems, 2)
		require.Equal(t, item1, soldItems[1])
		require.Equal(t, item2, soldItems[0])
	})
}
