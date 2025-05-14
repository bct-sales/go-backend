//go:build test

package queries

import (
	"bctbackend/algorithms"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestContainsFrozenItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("No frozen items", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false))
			itemIds := algorithms.Map(items, func(item *models.Item) models.Id { return item.ItemId })

			result, err := queries.ContainsFrozenItems(db, itemIds)
			require.NoError(t, err)
			require.False(t, result)
		})

		t.Run("One frozen item", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false))
			items = append(items, setup.Item(seller.UserId, aux.WithFrozen(true)))
			itemIds := algorithms.Map(items, func(item *models.Item) models.Id { return item.ItemId })

			result, err := queries.ContainsFrozenItems(db, itemIds)
			require.NoError(t, err)
			require.True(t, result)
		})

		t.Run("Duplicate items, no frozen items", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false))
			itemIds := algorithms.Map(items, func(item *models.Item) models.Id { return item.ItemId })
			itemIds = append(itemIds, itemIds...)

			result, err := queries.ContainsFrozenItems(db, itemIds)
			require.NoError(t, err)
			require.False(t, result)
		})

		t.Run("Duplicate items, frozen item", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false))
			items = append(items, setup.Item(seller.UserId, aux.WithFrozen(true)))
			itemIds := algorithms.Map(items, func(item *models.Item) models.Id { return item.ItemId })
			itemIds = append(itemIds, itemIds...)

			result, err := queries.ContainsFrozenItems(db, itemIds)
			require.NoError(t, err)
			require.True(t, result)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent item", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false))
			itemIds := algorithms.Map(items, func(item *models.Item) models.Id { return item.ItemId })
			nonexistentItemId := models.Id(1000)
			setup.RequireNoSuchItem(t, nonexistentItemId)
			itemIds = append(itemIds, nonexistentItemId)

			_, err := queries.ContainsFrozenItems(db, itemIds)
			var noSuchItemError *queries.NoSuchItemError
			require.ErrorAs(t, err, &noSuchItemError)
		})
	})
}
