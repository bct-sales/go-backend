//go:build test

package queries

import (
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestCountItems(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		count, err := queries.CountItems(db)
		require.NoError(t, err)
		require.Equal(t, 0, count)
	})

	t.Run("One item", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId
		AddItemToDatabase(db, sellerId, WithDummyData(1))

		count, err := queries.CountItems(db)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})

	t.Run("Two items", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId
		AddItemToDatabase(db, sellerId, WithDummyData(1))
		AddItemToDatabase(db, sellerId, WithDummyData(2))

		count, err := queries.CountItems(db)
		require.NoError(t, err)
		require.Equal(t, 2, count)
	})
}
