//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetSalesWithCashier(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Zero sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			cashier := setup.Cashier()
			sales, err := queries.GetSalesWithCashier(db, cashier.UserId)
			require.NoError(t, err)

			require.Empty(t, sales)
		})

		t.Run("One sale", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			cashier := setup.Cashier()
			seller := setup.Seller()

			item := setup.Item(seller.UserId, aux.WithDummyData(1))
			saleId := setup.Sale(cashier.UserId, []models.Id{item.ItemId})

			sales, err := queries.GetSalesWithCashier(db, cashier.UserId)
			require.NoError(t, err)

			require.Len(t, sales, 1)
			require.Equal(t, saleId, sales[0].SaleId)
			require.Equal(t, cashier.UserId, sales[0].CashierId)
		})

		t.Run("Two sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			cashier := setup.Cashier()
			seller := setup.Seller()

			item1 := setup.Item(seller.UserId, aux.WithDummyData(1))
			item2 := setup.Item(seller.UserId, aux.WithDummyData(2))
			saleId1 := setup.Sale(cashier.UserId, []models.Id{item1.ItemId})
			saleId2 := setup.Sale(cashier.UserId, []models.Id{item2.ItemId})

			sales, err := queries.GetSalesWithCashier(db, cashier.UserId)
			require.NoError(t, err)

			require.Len(t, sales, 2)
			require.Equal(t, saleId1, sales[0].SaleId)
			require.Equal(t, saleId2, sales[1].SaleId)
		})
	})
}
