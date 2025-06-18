//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContainsHiddenItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("No hidden items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
			itemIds := models.CollectItemIds(items)

			result, err := queries.ContainsHiddenItems(db, itemIds)
			require.NoError(t, err)
			require.False(t, result)
		})

		t.Run("One hidden item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
			items = append(items, setup.Item(seller.UserId, aux.WithHidden(true)))
			itemIds := models.CollectItemIds(items)

			result, err := queries.ContainsHiddenItems(db, itemIds)
			require.NoError(t, err)
			require.True(t, result)
		})

		t.Run("Duplicate items, no hidden items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
			itemIds := models.CollectItemIds(items)
			itemIds = append(itemIds, itemIds...)

			result, err := queries.ContainsHiddenItems(db, itemIds)
			require.NoError(t, err)
			require.False(t, result)
		})

		t.Run("Duplicate items, hidden item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
			items = append(items, setup.Item(seller.UserId, aux.WithHidden(true)))
			itemIds := models.CollectItemIds(items)
			itemIds = append(itemIds, itemIds...)

			result, err := queries.ContainsHiddenItems(db, itemIds)
			require.NoError(t, err)
			require.True(t, result)
		})
		t.Run("Nonexistent item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
			itemIds := models.CollectItemIds(items)
			nonexistentItemId := models.Id(1000)
			setup.RequireNoSuchItems(t, nonexistentItemId)
			itemIds = append(itemIds, nonexistentItemId)

			_, err := queries.ContainsHiddenItems(db, itemIds)
			require.NoError(t, err)
		})
	})
}
