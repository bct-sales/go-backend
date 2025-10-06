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
		t.Run("All sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			otherCashier := setup.Cashier()
			items := setup.Items(seller.UserId, 100, aux.WithHidden(false))

			sale1 := setup.Sale(cashier.UserId, []models.ID{items[0].ItemID})
			sale2 := setup.Sale(cashier.UserId, []models.ID{items[1].ItemID})
			setup.Sale(otherCashier.UserId, []models.ID{items[2].ItemID})

			actual := []*models.SaleSummary{}
			err := queries.GetCashierSales(db, cashier.UserId, queries.CollectTo(&actual), queries.OrderChronological, queries.AllRows())

			require.NoError(t, err)
			require.Len(t, actual, 2)
			require.Equal(t, sale1.SaleID, actual[0].SaleID)
			require.Equal(t, sale2.SaleID, actual[1].SaleID)
		})

		t.Run("Using row selection", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			items := setup.Items(seller.UserId, 100, aux.WithHidden(false))

			saleIds := []models.ID{}
			for _, item := range items {
				sale := setup.Sale(cashier.UserId, []models.ID{item.ItemID})
				saleIds = append(saleIds, sale.SaleID)
			}

			actual := []*models.SaleSummary{}
			err := queries.GetCashierSales(db, cashier.UserId, queries.CollectTo(&actual), queries.OrderChronological, queries.NewRowSelection(2, 5))

			require.NoError(t, err)
			require.Len(t, actual, 5)
			require.Equal(t, saleIds[2], actual[0].SaleID)
			require.Equal(t, saleIds[3], actual[1].SaleID)
			require.Equal(t, saleIds[4], actual[2].SaleID)
			require.Equal(t, saleIds[5], actual[3].SaleID)
			require.Equal(t, saleIds[6], actual[4].SaleID)
		})

		t.Run("Anti chronologically", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			items := setup.Items(seller.UserId, 3, aux.WithHidden(false))

			saleIds := []models.ID{}
			for _, item := range items {
				sale := setup.Sale(cashier.UserId, []models.ID{item.ItemID})
				saleIds = append(saleIds, sale.SaleID)
			}

			actual := []*models.SaleSummary{}
			err := queries.GetCashierSales(db, cashier.UserId, queries.CollectTo(&actual), queries.OrderAntiChronological, queries.AllRows())

			require.NoError(t, err)
			require.Len(t, actual, 3)
			require.Equal(t, saleIds[2], actual[0].SaleID)
			require.Equal(t, saleIds[1], actual[1].SaleID)
			require.Equal(t, saleIds[0], actual[2].SaleID)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent user", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			items := setup.Items(seller.UserId, 100, aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.ID{items[0].ItemID})
			setup.Sale(cashier.UserId, []models.ID{items[1].ItemID})

			cashierId := models.ID(9999)
			setup.RequireNoSuchUsers(t, cashierId)

			callback := func(sales *models.SaleSummary) error { return nil }
			err := queries.GetCashierSales(db, cashierId, callback, queries.OrderChronological, queries.AllRows())

			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
		})

		t.Run("Admin", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			admin := setup.Admin()
			items := setup.Items(seller.UserId, 100, aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.ID{items[0].ItemID})
			setup.Sale(cashier.UserId, []models.ID{items[1].ItemID})

			callback := func(sales *models.SaleSummary) error { return nil }
			err := queries.GetCashierSales(db, admin.UserId, callback, queries.OrderChronological, queries.AllRows())

			requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
		})

		t.Run("Seller", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			items := setup.Items(seller.UserId, 100, aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.ID{items[0].ItemID})
			setup.Sale(cashier.UserId, []models.ID{items[1].ItemID})

			callback := func(sales *models.SaleSummary) error { return nil }
			err := queries.GetCashierSales(db, seller.UserId, callback, queries.OrderChronological, queries.AllRows())

			requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
		})
	})
}
