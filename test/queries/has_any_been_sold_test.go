//go:build test

package queries

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHasAnyBeenSold(t *testing.T) {
	t.Run("Single unsold item", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()

		item := setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))

		actual, err := queries.HasAnyBeenSold(db, []models.ID{item.ItemID})
		require.NoError(t, err)
		require.False(t, actual)
	})

	t.Run("Single sold item", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()

		item := setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))
		setup.Sale(cashier.UserID, []models.ID{item.ItemID})

		actual, err := queries.HasAnyBeenSold(db, []models.ID{item.ItemID})
		require.NoError(t, err)
		require.True(t, actual)
	})

	t.Run("Multiple unsold items", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()

		items := setup.Items(seller.UserID, 10, aux.WithHidden(false))
		itemIDs := models.CollectItemIDs(items)

		actual, err := queries.HasAnyBeenSold(db, itemIDs)
		require.NoError(t, err)
		require.False(t, actual)
	})

	t.Run("Multiple unsold items", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()

		items := setup.Items(seller.UserID, 10, aux.WithHidden(false))
		itemIDs := models.CollectItemIDs(items)

		setup.Sale(cashier.UserID, itemIDs)

		actual, err := queries.HasAnyBeenSold(db, itemIDs)
		require.NoError(t, err)
		require.True(t, actual)
	})

	t.Run("Nonexistent item", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		nonexistentItemID := models.ID(1)
		setup.RequireNoSuchItems(t, nonexistentItemID)

		actual, err := queries.HasAnyBeenSold(db, []models.ID{nonexistentItemID})
		require.NoError(t, err)
		require.False(t, actual)
	})
}
