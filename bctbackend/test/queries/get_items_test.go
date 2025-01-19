//go:build test

package queries

import (
	"bctbackend/database/queries"
	"bctbackend/test"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetItems(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		items, err := queries.GetItems(db)
		require.NoError(t, err)
		require.Equal(t, 0, len(items))
	})

	t.Run("One item", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId
		itemId := test.AddItemToDatabase(db, sellerId, test.WithDummyData(1)).ItemId

		items, err := queries.GetItems(db)
		require.NoError(t, err)
		require.Equal(t, 1, len(items))
		require.Equal(t, itemId, items[0].ItemId)
	})

	t.Run("Two items", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId
		item1Id := test.AddItemToDatabase(db, sellerId, test.WithDummyData(1)).ItemId
		item2Id := test.AddItemToDatabase(db, sellerId, test.WithDummyData(2)).ItemId

		items, err := queries.GetItems(db)
		require.NoError(t, err)
		require.Equal(t, 2, len(items))
		require.Equal(t, item1Id, items[0].ItemId)
		require.Equal(t, item2Id, items[1].ItemId)
	})
}
