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

func TestRemoveItemWithId(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		itemId := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false)).ItemID

		err := queries.RemoveItemWithId(db, itemId)

		require.NoError(t, err)

		itemExists, err := queries.ItemWithIdExists(db, itemId)
		require.NoError(t, err)
		require.False(t, itemExists)
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			itemId := models.Id(1)

			err := queries.RemoveItemWithId(db, itemId)
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchItem)
		})

		t.Run("Sold item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			itemId := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false)).ItemID

			setup.Sale(cashier.UserId, []models.Id{itemId})

			err := queries.RemoveItemWithId(db, itemId)
			requireDatabaseWrappedError(t, err, dberr.ErrItemSold)

			itemExists, err := queries.ItemWithIdExists(db, itemId)
			require.NoError(t, err)
			require.True(t, itemExists)
		})
	})
}
