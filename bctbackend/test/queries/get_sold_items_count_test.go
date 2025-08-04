//go:build test

package queries

import (
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSoldItemsCount(t *testing.T) {
	t.Run("Only unsold items", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()

		setup.Items(seller.UserId, 10, aux.WithHidden(false))

		actual, err := queries.GetSoldItemsCount(db)
		require.NoError(t, err)
		require.Equal(t, 0, actual.Distinct)
		require.Equal(t, 0, actual.IncludeMultiples)
	})
}
