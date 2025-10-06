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
			items := setup.Items(seller.UserID, 10, aux.WithHidden(false))
			itemIDs := models.CollectItemIDs(items)

			result, err := queries.ContainsHiddenItems(db, itemIDs)
			require.NoError(t, err)
			require.False(t, result)
		})

		t.Run("One hidden item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserID, 10, aux.WithHidden(false))
			items = append(items, setup.Item(seller.UserID, aux.WithHidden(true)))
			itemIDs := models.CollectItemIDs(items)

			result, err := queries.ContainsHiddenItems(db, itemIDs)
			require.NoError(t, err)
			require.True(t, result)
		})

		t.Run("Duplicate items, no hidden items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserID, 10, aux.WithHidden(false))
			itemIDs := models.CollectItemIDs(items)
			itemIDs = append(itemIDs, itemIDs...)

			result, err := queries.ContainsHiddenItems(db, itemIDs)
			require.NoError(t, err)
			require.False(t, result)
		})

		t.Run("Duplicate items, hidden item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserID, 10, aux.WithHidden(false))
			items = append(items, setup.Item(seller.UserID, aux.WithHidden(true)))
			itemIDs := models.CollectItemIDs(items)
			itemIDs = append(itemIDs, itemIDs...)

			result, err := queries.ContainsHiddenItems(db, itemIDs)
			require.NoError(t, err)
			require.True(t, result)
		})
		t.Run("Nonexistent item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserID, 10, aux.WithHidden(false))
			itemIDs := models.CollectItemIDs(items)
			nonexistentItemID := setup.GenerateNonexistentItemID(t)
			itemIDs = append(itemIDs, nonexistentItemID)

			_, err := queries.ContainsHiddenItems(db, itemIDs)
			require.NoError(t, err)
		})
	})
}
