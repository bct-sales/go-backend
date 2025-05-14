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

func TestCountItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Empty", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			count, err := queries.CountItems(db)
			require.NoError(t, err)
			require.Equal(t, 0, count)
		})

		t.Run("One item", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			seller := setup.Seller()
			setup.Item(seller.UserId, aux.WithDummyData(1))

			count, err := queries.CountItems(db)
			require.NoError(t, err)
			require.Equal(t, 1, count)
		})

		t.Run("Two items", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			seller := setup.Seller()
			setup.Item(seller.UserId, aux.WithDummyData(1))
			setup.Item(seller.UserId, aux.WithDummyData(2))

			count, err := queries.CountItems(db)
			require.NoError(t, err)
			require.Equal(t, 2, count)
		})
	})
}
