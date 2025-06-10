//go:build test

package queries

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnsureItemsExist(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Visible items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
			itemIds := algorithms.Map(items, func(item *models.Item) models.Id { return item.ItemID })

			err := queries.EnsureItemsExist(db, itemIds)
			require.NoError(t, err)
		})

		t.Run("Hidden items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(true))
			itemIds := algorithms.Map(items, func(item *models.Item) models.Id { return item.ItemID })

			err := queries.EnsureItemsExist(db, itemIds)
			require.NoError(t, err)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		items := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
		nonexistentItemId := models.Id(150)
		setup.RequireNoSuchItems(t, nonexistentItemId)
		itemIds := append(algorithms.Map(items, func(item *models.Item) models.Id { return item.ItemID }), nonexistentItemId)

		err := queries.EnsureItemsExist(db, itemIds)
		require.ErrorIs(t, err, dberr.ErrNoSuchItem)
	})
}
