//go:build test

package queries

import (
	"bctbackend/algorithms"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestPartitionItemsByFrozenStatus(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("All items are visible, no nonexistent items in list", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			frozenItems := setup.Items(seller.UserId, 10, aux.WithFrozen(true), aux.WithHidden(false))
			unfrozenItems := setup.Items(seller.UserId, 5, aux.WithFrozen(false), aux.WithHidden(false))
			allItems := slices.Concat(
				algorithms.Map(frozenItems, func(i *models.Item) models.Id { return i.ItemId }),
				algorithms.Map(unfrozenItems, func(i *models.Item) models.Id { return i.ItemId }))

			actualUnfrozen, actualFrozen, err := queries.PartitionItemsByFrozenStatus(db, allItems)
			require.NoError(t, err)
			require.Equal(t, len(frozenItems), actualFrozen.Len())
			require.Equal(t, len(unfrozenItems), actualUnfrozen.Len())
		})

		t.Run("All items are hidden, no nonexistent items in list", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			frozenItems := setup.Items(seller.UserId, 10, aux.WithFrozen(true), aux.WithHidden(true))
			unfrozenItems := setup.Items(seller.UserId, 5, aux.WithFrozen(false), aux.WithHidden(true))
			allItems := slices.Concat(
				algorithms.Map(frozenItems, func(i *models.Item) models.Id { return i.ItemId }),
				algorithms.Map(unfrozenItems, func(i *models.Item) models.Id { return i.ItemId }))

			actualUnfrozen, actualFrozen, err := queries.PartitionItemsByFrozenStatus(db, allItems)
			require.NoError(t, err)
			require.Equal(t, len(frozenItems), actualFrozen.Len())
			require.Equal(t, len(unfrozenItems), actualUnfrozen.Len())
		})

		t.Run("All items are visible, nonexistent items in the list", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			frozenItems := setup.Items(seller.UserId, 10, aux.WithFrozen(true), aux.WithHidden(false))
			unfrozenItems := setup.Items(seller.UserId, 5, aux.WithFrozen(false), aux.WithHidden(false))
			nonexistentItems := []models.Id{999, 1000, 1001}
			// setup.RequireNoSuchItem(t, nonexistentItems)
			allItems := slices.Concat(
				algorithms.Map(frozenItems, func(i *models.Item) models.Id { return i.ItemId }),
				algorithms.Map(unfrozenItems, func(i *models.Item) models.Id { return i.ItemId }),
				nonexistentItems)

			actualUnfrozen, actualFrozen, err := queries.PartitionItemsByFrozenStatus(db, allItems)
			require.NoError(t, err)
			require.Equal(t, len(frozenItems), actualFrozen.Len())
			require.Equal(t, len(unfrozenItems), actualUnfrozen.Len())
		})

		t.Run("All items are hidden, nonexistent items in the list", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			frozenItems := setup.Items(seller.UserId, 10, aux.WithFrozen(true), aux.WithHidden(true))
			unfrozenItems := setup.Items(seller.UserId, 5, aux.WithFrozen(false), aux.WithHidden(true))
			nonexistentItems := []models.Id{999, 1000, 1001}
			// setup.RequireNoSuchItem(t, nonexistentItems)
			allItems := slices.Concat(
				algorithms.Map(frozenItems, func(i *models.Item) models.Id { return i.ItemId }),
				algorithms.Map(unfrozenItems, func(i *models.Item) models.Id { return i.ItemId }),
				nonexistentItems)

			actualUnfrozen, actualFrozen, err := queries.PartitionItemsByFrozenStatus(db, allItems)
			require.NoError(t, err)
			require.Equal(t, len(frozenItems), actualFrozen.Len())
			require.Equal(t, len(unfrozenItems), actualUnfrozen.Len())
		})
	})
}
