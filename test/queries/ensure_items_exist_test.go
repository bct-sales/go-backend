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
			itemIDs := models.CollectItemIDs(items)

			err := queries.EnsureItemsExist(db, itemIDs)
			require.NoError(t, err)
		})

		t.Run("Hidden items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(true))
			itemIDs := models.CollectItemIDs(items)

			err := queries.EnsureItemsExist(db, itemIDs)
			require.NoError(t, err)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		items := setup.Items(seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))
		nonexistentItemID := setup.GenerateNonexistentItemID(t)
		itemIDs := append(models.CollectItemIDs(items), nonexistentItemID)

		err := queries.EnsureItemsExist(db, itemIDs)
		requireDatabaseWrappedError(t, err, dberr.ErrNoSuchItem)
	})
}
