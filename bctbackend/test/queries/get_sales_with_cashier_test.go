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

func TestGetSalesWithCashier(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Zero sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier()
			sales, err := queries.GetSalesWithCashier(db, cashier.UserId)
			require.NoError(t, err)

			require.Empty(t, sales)
		})

		t.Run("One sale", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier()
			seller := setup.Seller()

			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
			saleId := setup.Sale(cashier.UserId, []models.Id{item.ItemID})

			sales, err := queries.GetSalesWithCashier(db, cashier.UserId)
			require.NoError(t, err)

			require.Len(t, sales, 1)
			require.Equal(t, saleId, sales[0].SaleID)
			require.Equal(t, cashier.UserId, sales[0].CashierID)
		})

		t.Run("Two sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier()
			seller := setup.Seller()

			item1 := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
			item2 := setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithHidden(false))
			saleId1 := setup.Sale(cashier.UserId, []models.Id{item1.ItemID})
			saleId2 := setup.Sale(cashier.UserId, []models.Id{item2.ItemID})

			sales, err := queries.GetSalesWithCashier(db, cashier.UserId)
			require.NoError(t, err)

			require.Len(t, sales, 2)
			require.Equal(t, saleId1, sales[0].SaleID)
			require.Equal(t, saleId2, sales[1].SaleID)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Unknown cashier", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			unknownCashierId := models.Id(9999)
			setup.RequireNoSuchUsers(t, unknownCashierId)

			_, err := queries.GetSalesWithCashier(db, unknownCashierId)
			require.ErrorIs(t, err, dberr.ErrNoSuchUser)
		})

		t.Run("User whose sales we want is an admin", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			admin := setup.Admin()

			_, err := queries.GetSalesWithCashier(db, admin.UserId)
			require.ErrorIs(t, err, dberr.ErrWrongRole)
		})

		t.Run("User whose sales we want is a seller", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()

			_, err := queries.GetSalesWithCashier(db, seller.UserId)
			require.ErrorIs(t, err, dberr.ErrWrongRole)
		})
	})
}
