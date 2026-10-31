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

func TestEnsureNoHiddenItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)
		items := setup.Items(t, seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))
		itemIDs := models.CollectItemIDs(items)

		err := queries.EnsureNoHiddenItems(t.Context(), db, itemIDs)
		require.NoError(t, err)
	})

	t.Run("Failure", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)
		visibleItems := setup.Items(t, seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))
		hiddenItem := setup.Item(t, seller.UserID, aux.WithFrozen(false), aux.WithHidden(true))
		itemIDs := append(models.CollectItemIDs(visibleItems), hiddenItem.ItemID)

		err := queries.EnsureNoHiddenItems(t.Context(), db, itemIDs)
		requireDatabaseWrappedError(t, err, dberr.ErrItemHidden)
	})
}
