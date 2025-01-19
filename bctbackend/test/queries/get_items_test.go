//go:build test

package queries

import (
	"bctbackend/database/queries"
	"bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetItems(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		db := setup.OpenInitializedDatabase()
		defer db.Close()

		items, err := queries.GetItems(db)
		require.NoError(t, err)
		require.Equal(t, 0, len(items))
	})

	t.Run("One item", func(t *testing.T) {
		db := setup.OpenInitializedDatabase()
		defer db.Close()

		sellerId := setup.AddSellerToDatabase(db).UserId
		itemId := setup.AddItemToDatabase(db, sellerId, setup.WithDummyData(1)).ItemId

		items, err := queries.GetItems(db)
		require.NoError(t, err)
		require.Equal(t, 1, len(items))
		require.Equal(t, itemId, items[0].ItemId)
	})

	t.Run("Two items", func(t *testing.T) {
		db := setup.OpenInitializedDatabase()
		defer db.Close()

		sellerId := setup.AddSellerToDatabase(db).UserId
		item1Id := setup.AddItemToDatabase(db, sellerId, setup.WithDummyData(1)).ItemId
		item2Id := setup.AddItemToDatabase(db, sellerId, setup.WithDummyData(2)).ItemId

		items, err := queries.GetItems(db)
		require.NoError(t, err)
		require.Equal(t, 2, len(items))
		require.Equal(t, item1Id, items[0].ItemId)
		require.Equal(t, item2Id, items[1].ItemId)
	})
}
