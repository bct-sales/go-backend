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
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		items := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
		itemIds := models.CollectItemIds(items)

		err := queries.EnsureNoHiddenItems(db, itemIds)
		require.NoError(t, err)
	})

	t.Run("Failure", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		visibleItems := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
		hiddenItem := setup.Item(seller.UserId, aux.WithFrozen(false), aux.WithHidden(true))
		itemIds := append(models.CollectItemIds(visibleItems), hiddenItem.ItemID)

		err := queries.EnsureNoHiddenItems(db, itemIds)
		require.ErrorIs(t, err, dberr.ErrItemHidden)
	})
}
