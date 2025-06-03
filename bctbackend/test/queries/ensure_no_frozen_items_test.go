//go:build test

package queries

import (
	"bctbackend/algorithms"
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureNoFrozenItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		items := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
		itemIds := algorithms.Map(items, func(item *models.Item) models.Id { return item.ItemId })

		err := queries.EnsureNoFrozenItems(db, itemIds)
		require.NoError(t, err)
	})

	t.Run("Failure", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		unfrozenItems := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
		frozenItem := setup.Item(seller.UserId, aux.WithFrozen(true), aux.WithHidden(false))
		itemIds := append(algorithms.Map(unfrozenItems, func(item *models.Item) models.Id { return item.ItemId }), frozenItem.ItemId)

		err := queries.EnsureNoFrozenItems(db, itemIds)
		require.ErrorIs(t, err, database.ErrItemFrozen)
	})
}
