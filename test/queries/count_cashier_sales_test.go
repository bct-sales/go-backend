//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCountCashierSales(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Multiple sales, single cashier", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()

			items := setup.Items(seller.UserID, 5, aux.WithHidden(false))
			setup.Sale(cashier.UserID, []models.ID{items[0].ItemID, items[1].ItemID})
			setup.Sale(cashier.UserID, []models.ID{items[2].ItemID, items[3].ItemID})
			setup.Sale(cashier.UserID, []models.ID{items[4].ItemID})

			count, err := queries.CountCashierSales(db, cashier.UserID)
			require.NoError(t, err)
			require.Equal(t, 3, count)
		})

		t.Run("Multiple sales, multiple cashiers", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			cashier2 := setup.Cashier()

			items := setup.Items(seller.UserID, 5, aux.WithHidden(false))
			setup.Sale(cashier.UserID, []models.ID{items[0].ItemID, items[1].ItemID})
			setup.Sale(cashier.UserID, []models.ID{items[2].ItemID, items[3].ItemID})
			setup.Sale(cashier2.UserID, []models.ID{items[4].ItemID})

			count, err := queries.CountCashierSales(db, cashier.UserID)
			require.NoError(t, err)
			require.Equal(t, 2, count)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent cashier", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			nonexistentCashierId := models.ID(999)
			setup.RequireNoSuchUsers(t, nonexistentCashierId)

			_, err := queries.CountCashierSales(db, nonexistentCashierId)
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
		})

		t.Run("Not a cashier", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			noncashier := setup.Seller()

			_, err := queries.CountCashierSales(db, noncashier.UserID)
			requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
		})
	})
}
