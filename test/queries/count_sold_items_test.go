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
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()

		setup.Items(seller.UserId, 10, aux.WithHidden(false))

		actual, err := queries.CountSoldItems(db)
		require.NoError(t, err)
		require.Equal(t, 0, actual.Distinct)
		require.Equal(t, 0, actual.IncludeMultiples)
	})

	t.Run("Sales without overlaps", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()

		items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
		setup.Sale(cashier.UserId, []models.ID{items[0].ItemID})
		setup.Sale(cashier.UserId, []models.ID{items[1].ItemID, items[2].ItemID})

		actual, err := queries.CountSoldItems(db)
		require.NoError(t, err)
		require.Equal(t, 3, actual.Distinct)
		require.Equal(t, 3, actual.IncludeMultiples)
	})

	t.Run("Sales with shared items", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()

		items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
		setup.Sale(cashier.UserId, []models.ID{items[0].ItemID})
		setup.Sale(cashier.UserId, []models.ID{items[0].ItemID, items[2].ItemID})

		actual, err := queries.CountSoldItems(db)
		require.NoError(t, err)
		require.Equal(t, 2, actual.Distinct)
		require.Equal(t, 3, actual.IncludeMultiples)
	})
}
