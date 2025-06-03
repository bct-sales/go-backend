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

func TestEnsureNoHiddenItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		items := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
		itemIds := algorithms.Map(items, func(item *models.Item) models.Id { return item.ItemId })

		err := queries.EnsureNoHiddenItems(db, itemIds)
		require.NoError(t, err)
	})

	t.Run("Failure", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		visibleItems := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
		hiddenItem := setup.Item(seller.UserId, aux.WithFrozen(false), aux.WithHidden(true))
		itemIds := append(algorithms.Map(visibleItems, func(item *models.Item) models.Id { return item.ItemId }), hiddenItem.ItemId)

		err := queries.EnsureNoHiddenItems(db, itemIds)
		require.ErrorIs(t, err, database.ErrItemHidden)
	})
}
