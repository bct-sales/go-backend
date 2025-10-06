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
)

func TestPartitionItemsByFrozenStatus(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("All items are visible, no nonexistent items in list", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			frozenItems := setup.Items(seller.UserID, 10, aux.WithFrozen(true), aux.WithHidden(false))
			unfrozenItems := setup.Items(seller.UserID, 5, aux.WithFrozen(false), aux.WithHidden(false))
			allItems := slices.Concat(
				algorithms.Map(frozenItems, func(i *models.Item) models.ID { return i.ItemID }),
				algorithms.Map(unfrozenItems, func(i *models.Item) models.ID { return i.ItemID }))

			actualUnfrozen, actualFrozen, err := queries.PartitionItemsByFrozenStatus(db, allItems)
			require.NoError(t, err)
			require.Equal(t, len(frozenItems), actualFrozen.Len())
			require.Equal(t, len(unfrozenItems), actualUnfrozen.Len())
		})

		t.Run("Some unfrozen items are hidden, no nonexistent items in list", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			frozenItems := setup.Items(seller.UserID, 10, aux.WithFrozen(true), aux.WithHidden(false))
			unfrozenVisibleItems := setup.Items(seller.UserID, 5, aux.WithFrozen(false), aux.WithHidden(false))
			unfrozenHiddenItems := setup.Items(seller.UserID, 5, aux.WithFrozen(false), aux.WithHidden(true))
			allItems := slices.Concat(
				algorithms.Map(frozenItems, func(i *models.Item) models.ID { return i.ItemID }),
				algorithms.Map(unfrozenVisibleItems, func(i *models.Item) models.ID { return i.ItemID }),
				algorithms.Map(unfrozenHiddenItems, func(i *models.Item) models.ID { return i.ItemID }),
			)

			actualUnfrozen, actualFrozen, err := queries.PartitionItemsByFrozenStatus(db, allItems)
			require.NoError(t, err)
			require.Equal(t, len(frozenItems), actualFrozen.Len())
			require.Equal(t, len(unfrozenVisibleItems)+len(unfrozenHiddenItems), actualUnfrozen.Len())
		})

		t.Run("All items are visible, nonexistent items in the list", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			frozenItems := setup.Items(seller.UserID, 10, aux.WithFrozen(true), aux.WithHidden(false))
			unfrozenItems := setup.Items(seller.UserID, 5, aux.WithFrozen(false), aux.WithHidden(false))
			nonexistentItems := []models.ID{
				setup.GenerateNonexistentItemID(t),
				setup.GenerateNonexistentItemID(t),
				setup.GenerateNonexistentItemID(t),
			}
			allItems := slices.Concat(
				algorithms.Map(frozenItems, func(i *models.Item) models.ID { return i.ItemID }),
				algorithms.Map(unfrozenItems, func(i *models.Item) models.ID { return i.ItemID }),
				nonexistentItems)

			actualUnfrozen, actualFrozen, err := queries.PartitionItemsByFrozenStatus(db, allItems)
			require.NoError(t, err)
			require.Equal(t, len(frozenItems), actualFrozen.Len())
			require.Equal(t, len(unfrozenItems), actualUnfrozen.Len())
		})
	})
}
