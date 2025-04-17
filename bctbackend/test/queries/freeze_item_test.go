//go:build test

package queries

import (
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestFreezeItem(t *testing.T) {
	t.Run("Successful freezing", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		seller := setup.Seller()
		item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithFrozen(false))

		{
			isFrozen, err := queries.ItemWithIdIsFrozen(db, item.ItemId)
			require.NoError(t, err)
			require.False(t, isFrozen)
		}

		{
			err := queries.FreezeItem(db, item.ItemId)
			require.NoError(t, err)
		}

		{
			isFrozen, err := queries.ItemWithIdIsFrozen(db, item.ItemId)
			require.NoError(t, err)
			require.True(t, isFrozen)
		}
	})
}
