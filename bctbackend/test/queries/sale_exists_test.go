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
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			itemId := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false)).ItemID

			sale := setup.Sale(cashier.UserId, []models.Id{itemId})
			saleExists, err := queries.SaleWithIdExists(db, sale.SaleID)
			require.NoError(t, err)
			require.True(t, saleExists)
		})

		t.Run("Sale does not exist", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			saleExists, err := queries.SaleWithIdExists(db, 1)
			require.NoError(t, err)
			require.False(t, saleExists)
		})
	})
}
