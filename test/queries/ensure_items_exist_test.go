//go:build test

package queries

import (
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
			items := setup.Items(seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))
			itemIds := models.CollectItemIDs(items)

			err := queries.EnsureItemsExist(db, itemIds)
			require.NoError(t, err)
		})

		t.Run("Hidden items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(true))
			itemIds := models.CollectItemIDs(items)

			err := queries.EnsureItemsExist(db, itemIds)
			require.NoError(t, err)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		items := setup.Items(seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))
		nonexistentItemId := models.ID(150)
		setup.RequireNoSuchItems(t, nonexistentItemId)
		itemIds := append(models.CollectItemIDs(items), nonexistentItemId)

		err := queries.EnsureItemsExist(db, itemIds)
		requireDatabaseWrappedError(t, err, dberr.ErrNoSuchItem)
	})
}
