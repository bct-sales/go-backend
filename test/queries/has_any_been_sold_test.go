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

		item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))

		actual, err := queries.HasAnyBeenSold(db, []models.Id{item.ItemID})
		require.NoError(t, err)
		require.False(t, actual)
	})

	t.Run("Single sold item", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()

		item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
		setup.Sale(cashier.UserId, []models.Id{item.ItemID})

		actual, err := queries.HasAnyBeenSold(db, []models.Id{item.ItemID})
		require.NoError(t, err)
		require.True(t, actual)
	})

	t.Run("Multiple unsold items", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()

		items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
		itemIds := models.CollectItemIds(items)

		actual, err := queries.HasAnyBeenSold(db, itemIds)
		require.NoError(t, err)
		require.False(t, actual)
	})

	t.Run("Multiple unsold items", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()

		items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
		itemIds := models.CollectItemIds(items)

		setup.Sale(cashier.UserId, itemIds)

		actual, err := queries.HasAnyBeenSold(db, itemIds)
		require.NoError(t, err)
		require.True(t, actual)
	})

	t.Run("Nonexistent item", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		nonexistentItemId := models.Id(1)
		setup.RequireNoSuchItems(t, nonexistentItemId)

		actual, err := queries.HasAnyBeenSold(db, []models.Id{nonexistentItemId})
		require.NoError(t, err)
		require.False(t, actual)
	})
}
