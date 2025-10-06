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

func TestPartitionItemsByHiddenStatus(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("No nonexistent items in list", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			visibleItems := setup.Items(seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))
			hiddenItems := setup.Items(seller.UserID, 5, aux.WithFrozen(false), aux.WithHidden(true))
			allItems := slices.Concat(
				algorithms.Map(visibleItems, func(i *models.Item) models.ID { return i.ItemID }),
				algorithms.Map(hiddenItems, func(i *models.Item) models.ID { return i.ItemID }))

			actualVisible, actualHidden, err := queries.PartitionItemsByHiddenStatus(db, allItems)
			require.NoError(t, err)
			require.Equal(t, len(visibleItems), actualVisible.Len())
			require.Equal(t, len(hiddenItems), actualHidden.Len())
		})

		t.Run("All items are hidden, no nonexistent items in list", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			visibleItems := setup.Items(seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))
			hiddenItems := setup.Items(seller.UserID, 5, aux.WithFrozen(false), aux.WithHidden(true))
			nonexistentItemIDs := []models.ID{setup.GenerateNonexistentItemID(t), setup.GenerateNonexistentItemID(t)}
			allItems := slices.Concat(
				algorithms.Map(visibleItems, func(i *models.Item) models.ID { return i.ItemID }),
				algorithms.Map(hiddenItems, func(i *models.Item) models.ID { return i.ItemID }),
				nonexistentItemIDs)

			actualVisible, actualHidden, err := queries.PartitionItemsByHiddenStatus(db, allItems)
			require.NoError(t, err)
			require.Equal(t, len(visibleItems), actualVisible.Len())
			require.Equal(t, len(hiddenItems), actualHidden.Len())
		})
	})
}
