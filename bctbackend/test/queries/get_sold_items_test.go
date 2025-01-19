//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/test"
	"bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetSoldItems(t *testing.T) {
	t.Run("No items in existence", func(t *testing.T) {
		db := setup.OpenInitializedDatabase()
		defer db.Close()

		soldItems, err := queries.GetSoldItems(db)
		require.NoError(t, err)
		require.Empty(t, soldItems)
	})

	t.Run("Single unsold item", func(t *testing.T) {
		db := setup.OpenInitializedDatabase()
		defer db.Close()

		seller := test.AddSellerToDatabase(db)
		test.AddItemToDatabase(db, seller.UserId, test.WithDummyData(1))

		soldItems, err := queries.GetSoldItems(db)
		require.NoError(t, err)
		require.Empty(t, soldItems)
	})

	t.Run("Single sold item", func(t *testing.T) {
		db := setup.OpenInitializedDatabase()
		defer db.Close()

		seller := test.AddSellerToDatabase(db)
		cashier := test.AddCashierToDatabase(db)
		item := test.AddItemToDatabase(db, seller.UserId, test.WithDummyData(1))
		test.AddSaleToDatabase(db, cashier.UserId, []models.Id{item.ItemId})

		soldItems, err := queries.GetSoldItems(db)
		require.NoError(t, err)
		require.Len(t, soldItems, 1)
		require.Equal(t, item, &soldItems[0])
	})

	t.Run("Doubly sold item", func(t *testing.T) {
		db := setup.OpenInitializedDatabase()
		defer db.Close()

		seller := test.AddSellerToDatabase(db)
		cashier := test.AddCashierToDatabase(db)
		item := test.AddItemToDatabase(db, seller.UserId, test.WithDummyData(1))
		test.AddSaleToDatabase(db, cashier.UserId, []models.Id{item.ItemId})
		test.AddSaleToDatabase(db, cashier.UserId, []models.Id{item.ItemId})

		soldItems, err := queries.GetSoldItems(db)
		require.NoError(t, err)
		require.Len(t, soldItems, 1)
		require.Equal(t, item, &soldItems[0])
	})

	t.Run("Two sold items in single sale", func(t *testing.T) {
		db := setup.OpenInitializedDatabase()
		defer db.Close()

		seller := test.AddSellerToDatabase(db)
		cashier := test.AddCashierToDatabase(db)
		item1 := test.AddItemToDatabase(db, seller.UserId, test.WithDummyData(1))
		item2 := test.AddItemToDatabase(db, seller.UserId, test.WithDummyData(2))
		test.AddSaleToDatabase(db, cashier.UserId, []models.Id{item1.ItemId, item2.ItemId})

		soldItems, err := queries.GetSoldItems(db)
		require.NoError(t, err)
		require.Len(t, soldItems, 2)
		require.Equal(t, item1, &soldItems[0])
		require.Equal(t, item2, &soldItems[1])
	})

	t.Run("Two sold items in separate sales", func(t *testing.T) {
		db := setup.OpenInitializedDatabase()
		defer db.Close()

		seller := test.AddSellerToDatabase(db)
		cashier := test.AddCashierToDatabase(db)
		item1 := test.AddItemToDatabase(db, seller.UserId, test.WithDummyData(1))
		item2 := test.AddItemToDatabase(db, seller.UserId, test.WithDummyData(2))
		test.AddSaleAtTimeToDatabase(db, cashier.UserId, []models.Id{item1.ItemId}, models.Timestamp(100))
		test.AddSaleAtTimeToDatabase(db, cashier.UserId, []models.Id{item2.ItemId}, models.Timestamp(200))

		soldItems, err := queries.GetSoldItems(db)
		require.NoError(t, err)
		require.Len(t, soldItems, 2)
		require.Equal(t, item1, &soldItems[1])
		require.Equal(t, item2, &soldItems[0])
	})
}
