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

func TestSaleExists(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Sale exists", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)
			cashier := setup.Cashier(t)
			itemID := setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithHidden(false)).ItemID

			sale := setup.Sale(t, cashier.UserID, []models.ID{itemID})
			saleExists, err := queries.SaleWithIDExists(t.Context(), db, sale.SaleID)
			require.NoError(t, err)
			require.True(t, saleExists)
		})

		t.Run("Sale does not exist", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			saleExists, err := queries.SaleWithIDExists(t.Context(), db, 1)
			require.NoError(t, err)
			require.False(t, saleExists)
		})
	})
}
