//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test"
	aux "bctbackend/test/helpers"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetItems(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		items := []*models.Item{}
		err := queries.GetItems(db, queries.CollectTo(&items))
		require.NoError(t, err)
		require.Equal(t, 0, len(items))
	})

	t.Run("One item", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		seller := setup.Seller()
		item := setup.Item(seller.UserId, aux.WithDummyData(1))

		items := []*models.Item{}
		err := queries.GetItems(db, queries.CollectTo(&items))
		require.NoError(t, err)
		require.Equal(t, 1, len(items))
		require.Equal(t, item.ItemId, items[0].ItemId)
	})

	t.Run("Two items", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		seller := setup.Seller()
		item1 := setup.Item(seller.UserId, aux.WithDummyData(1))
		item2 := setup.Item(seller.UserId, aux.WithDummyData(2))

		items := []*models.Item{}
		err := queries.GetItems(db, queries.CollectTo(&items))
		require.NoError(t, err)
		require.Equal(t, 2, len(items))
		require.Equal(t, item1.ItemId, items[0].ItemId)
		require.Equal(t, item2.ItemId, items[1].ItemId)
	})
}
