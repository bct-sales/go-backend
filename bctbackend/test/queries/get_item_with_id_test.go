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

func TestGetItemWithId(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))

		actual, err := queries.GetItemWithId(db, item.ItemID)
		require.NoError(t, err)
		require.Equal(t, item, actual)
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexisting item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			itemId := models.Id(1)
			_, err := queries.GetItemWithId(db, itemId)
			require.ErrorIs(t, err, dberr.ErrNoSuchItem)
		})
	})
}
