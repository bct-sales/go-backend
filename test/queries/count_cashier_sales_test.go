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

func TestCountCashierSales(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Multiple sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()

			items := setup.Items(seller.UserId, 5, aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.Id{items[0].ItemID, items[1].ItemID})
			setup.Sale(cashier.UserId, []models.Id{items[2].ItemID, items[3].ItemID})
			setup.Sale(cashier.UserId, []models.Id{items[4].ItemID})

			count, err := queries.CountCashierSales(db, cashier.UserId)
			require.NoError(t, err)
			require.Equal(t, 3, count)
		})
	})
}
