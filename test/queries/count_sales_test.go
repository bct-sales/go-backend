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

func TestCountSales(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Zero sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)
			setup.Cashier(t)

			setup.Items(t, seller.UserID, 5, aux.WithHidden(false))

			count, err := queries.CountSales(t.Context(), db)
			require.NoError(t, err)
			require.Equal(t, 0, count)
		})

		t.Run("Single sale", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)
			cashier := setup.Cashier(t)

			items := setup.Items(t, seller.UserID, 5, aux.WithHidden(false))
			setup.Sale(t, cashier.UserID, []models.ID{items[0].ItemID, items[1].ItemID})

			count, err := queries.CountSales(t.Context(), db)
			require.NoError(t, err)
			require.Equal(t, 1, count)
		})

		t.Run("Multiple sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)
			cashier := setup.Cashier(t)

			items := setup.Items(t, seller.UserID, 5, aux.WithHidden(false))
			setup.Sale(t, cashier.UserID, []models.ID{items[0].ItemID, items[1].ItemID})
			setup.Sale(t, cashier.UserID, []models.ID{items[2].ItemID, items[3].ItemID})
			setup.Sale(t, cashier.UserID, []models.ID{items[4].ItemID})

			count, err := queries.CountSales(t.Context(), db)
			require.NoError(t, err)
			require.Equal(t, 3, count)
		})
	})
}
