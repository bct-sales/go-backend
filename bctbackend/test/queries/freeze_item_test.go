//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestFreezeItem(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Freezing unfrozen item", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			seller := setup.Seller()
			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithFrozen(false))

			{
				err := queries.FreezeItem(db, item.ItemId)
				require.NoError(t, err)
			}

			{
				isFrozen, err := queries.IsItemFrozen(db, item.ItemId)
				require.NoError(t, err)
				require.True(t, isFrozen)
			}
		})

		t.Run("Freezing already unfrozen item", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			seller := setup.Seller()
			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithFrozen(true))

			{
				err := queries.FreezeItem(db, item.ItemId)
				require.NoError(t, err)
			}

			{
				isFrozen, err := queries.IsItemFrozen(db, item.ItemId)
				require.NoError(t, err)
				require.True(t, isFrozen)
			}
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Freezing non-existent item", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			nonexistentItemId := models.Id(1000)
			setup.RequireNoSuchItem(t, nonexistentItemId)

			err := queries.FreezeItem(db, nonexistentItemId)
			var noSuchItemError *queries.NoSuchItemError
			require.ErrorAs(t, err, &noSuchItemError)
		})
	})
}
