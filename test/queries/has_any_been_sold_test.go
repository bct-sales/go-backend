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
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)

		item := setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))

		actual, err := queries.HasAnyBeenSold(t.Context(), db, []models.ID{item.ItemID})
		require.NoError(t, err)
		require.False(t, actual)
	})

	t.Run("Single sold item", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)
		cashier := setup.Cashier(t)

		item := setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))
		setup.Sale(t, cashier.UserID, []models.ID{item.ItemID})

		actual, err := queries.HasAnyBeenSold(t.Context(), db, []models.ID{item.ItemID})
		require.NoError(t, err)
		require.True(t, actual)
	})

	t.Run("Multiple unsold items", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)

		items := setup.Items(t, seller.UserID, 10, aux.WithHidden(false))
		itemIDs := models.CollectItemIDs(items)

		actual, err := queries.HasAnyBeenSold(t.Context(), db, itemIDs)
		require.NoError(t, err)
		require.False(t, actual)
	})

	t.Run("Multiple unsold items", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)
		cashier := setup.Cashier(t)

		items := setup.Items(t, seller.UserID, 10, aux.WithHidden(false))
		itemIDs := models.CollectItemIDs(items)

		setup.Sale(t, cashier.UserID, itemIDs)

		actual, err := queries.HasAnyBeenSold(t.Context(), db, itemIDs)
		require.NoError(t, err)
		require.True(t, actual)
	})

	t.Run("Nonexistent item", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		nonexistentItemID := setup.GenerateNonexistentItemID(t)

		actual, err := queries.HasAnyBeenSold(t.Context(), db, []models.ID{nonexistentItemID})
		require.NoError(t, err)
		require.False(t, actual)
	})
}
