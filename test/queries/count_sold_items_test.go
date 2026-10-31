//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCountSoldItems(t *testing.T) {
	t.Run("Only unsold items", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)

		setup.Items(t, seller.UserID, 10, aux.WithHidden(false))

		actual, err := queries.CountSoldItems(t.Context(), db)
		require.NoError(t, err)
		require.Equal(t, 0, actual.Distinct)
		require.Equal(t, 0, actual.IncludeMultiples)
	})

	t.Run("Sales without overlaps", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)
		cashier := setup.Cashier(t)

		items := setup.Items(t, seller.UserID, 10, aux.WithHidden(false))
		setup.Sale(t, cashier.UserID, []models.ID{items[0].ItemID})
		setup.Sale(t, cashier.UserID, []models.ID{items[1].ItemID, items[2].ItemID})

		actual, err := queries.CountSoldItems(t.Context(), db)
		require.NoError(t, err)
		require.Equal(t, 3, actual.Distinct)
		require.Equal(t, 3, actual.IncludeMultiples)
	})

	t.Run("Sales with shared items", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)
		cashier := setup.Cashier(t)

		items := setup.Items(t, seller.UserID, 10, aux.WithHidden(false))
		setup.Sale(t, cashier.UserID, []models.ID{items[0].ItemID})
		setup.Sale(t, cashier.UserID, []models.ID{items[0].ItemID, items[2].ItemID})

		actual, err := queries.CountSoldItems(t.Context(), db)
		require.NoError(t, err)
		require.Equal(t, 2, actual.Distinct)
		require.Equal(t, 3, actual.IncludeMultiples)
	})
}
