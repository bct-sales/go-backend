//go:build test

package queries

import (
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestFreezeItem(t *testing.T) {
	t.Run("Successful freezing", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		seller := AddSellerToDatabase(db)
		item := AddItemToDatabase(db, seller.UserId, WithDummyData(1), WithFrozen(false))

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
