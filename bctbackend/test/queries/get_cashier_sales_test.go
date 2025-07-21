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

func TestGetCashierSales(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()
		otherCashier := setup.Cashier()
		items := setup.Items(seller.UserId, 100, aux.WithHidden(false))

		sale1 := setup.Sale(cashier.UserId, []models.Id{items[0].ItemID})
		sale2 := setup.Sale(cashier.UserId, []models.Id{items[1].ItemID})
		setup.Sale(otherCashier.UserId, []models.Id{items[2].ItemID})

		actual := []*models.SaleSummary{}
		err := queries.GetCashierSales(db, cashier.UserId, queries.CollectTo(&actual))

		require.NoError(t, err)
		require.Len(t, actual, 2)
		require.Equal(t, sale1.SaleID, actual[0].SaleID)
		require.Equal(t, sale2.SaleID, actual[1].SaleID)
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent user", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			items := setup.Items(seller.UserId, 100, aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.Id{items[0].ItemID})
			setup.Sale(cashier.UserId, []models.Id{items[1].ItemID})

			cashierId := models.Id(9999)
			setup.RequireNoSuchUsers(t, cashierId)

			callback := func(sales *models.SaleSummary) error { return nil }
			err := queries.GetCashierSales(db, cashierId, callback)

			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
		})

		t.Run("Admin", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			admin := setup.Admin()
			items := setup.Items(seller.UserId, 100, aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.Id{items[0].ItemID})
			setup.Sale(cashier.UserId, []models.Id{items[1].ItemID})

			callback := func(sales *models.SaleSummary) error { return nil }
			err := queries.GetCashierSales(db, admin.UserId, callback)

			requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
		})

		t.Run("Seller", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			items := setup.Items(seller.UserId, 100, aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.Id{items[0].ItemID})
			setup.Sale(cashier.UserId, []models.Id{items[1].ItemID})

			callback := func(sales *models.SaleSummary) error { return nil }
			err := queries.GetCashierSales(db, seller.UserId, callback)

			requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
		})
	})
}
