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
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier(t)
			sales, err := queries.GetSalesWithCashier(t.Context(), db, cashier.UserID)
			require.NoError(t, err)

			require.Empty(t, sales)
		})

		t.Run("One sale", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier(t)
			seller := setup.Seller(t)

			item := setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))
			sale := setup.Sale(t, cashier.UserID, []models.ID{item.ItemID})

			sales, err := queries.GetSalesWithCashier(t.Context(), db, cashier.UserID)
			require.NoError(t, err)

			require.Len(t, sales, 1)
			require.Equal(t, sale.SaleID, sales[0].SaleID)
			require.Equal(t, cashier.UserID, sales[0].CashierID)
		})

		t.Run("Two sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier(t)
			seller := setup.Seller(t)

			item1 := setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))
			item2 := setup.Item(t, seller.UserID, aux.WithDummyData(2), aux.WithHidden(false))
			sale1 := setup.Sale(t, cashier.UserID, []models.ID{item1.ItemID})
			sale2 := setup.Sale(t, cashier.UserID, []models.ID{item2.ItemID})

			sales, err := queries.GetSalesWithCashier(t.Context(), db, cashier.UserID)
			require.NoError(t, err)

			require.Len(t, sales, 2)
			require.Equal(t, sale1.SaleID, sales[0].SaleID)
			require.Equal(t, sale2.SaleID, sales[1].SaleID)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Unknown cashier", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			unknownCashierID := setup.GenerateNonexistentUserID(t)

			_, err := queries.GetSalesWithCashier(t.Context(), db, unknownCashierID)
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
		})

		t.Run("User whose sales we want is an admin", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			admin := setup.Admin(t)

			_, err := queries.GetSalesWithCashier(t.Context(), db, admin.UserID)
			requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
		})

		t.Run("User whose sales we want is a seller", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)

			_, err := queries.GetSalesWithCashier(t.Context(), db, seller.UserID)
			requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
		})
	})
}
