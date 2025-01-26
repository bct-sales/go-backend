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

func TestGetItems(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		items := []*models.Item{}
		err := queries.GetItems(db, queries.CollectTo(&items))
		require.NoError(t, err)
		require.Equal(t, 0, len(items))
	})

	t.Run("One item", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId
		itemId := AddItemToDatabase(db, sellerId, WithDummyData(1)).ItemId

		items := []*models.Item{}
		err := queries.GetItems(db, queries.CollectTo(&items))
		require.NoError(t, err)
		require.Equal(t, 1, len(items))
		require.Equal(t, itemId, items[0].ItemId)
	})

	t.Run("Two items", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId
		item1Id := AddItemToDatabase(db, sellerId, WithDummyData(1)).ItemId
		item2Id := AddItemToDatabase(db, sellerId, WithDummyData(2)).ItemId

		items := []*models.Item{}
		err := queries.GetItems(db, queries.CollectTo(&items))
		require.NoError(t, err)
		require.Equal(t, 2, len(items))
		require.Equal(t, item1Id, items[0].ItemId)
		require.Equal(t, item2Id, items[1].ItemId)
	})
}
