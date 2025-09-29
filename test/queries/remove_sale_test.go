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

func TestRemoveSale(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()
		sale1ItemIds := []models.Id{
			setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false)).ItemID,
			setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithHidden(false)).ItemID,
		}
		sale2ItemIds := []models.Id{
			setup.Item(seller.UserId, aux.WithDummyData(3), aux.WithHidden(false)).ItemID,
			setup.Item(seller.UserId, aux.WithDummyData(4), aux.WithHidden(false)).ItemID,
		}

		sale1 := setup.Sale(cashier.UserId, sale1ItemIds)
		sale2 := setup.Sale(cashier.UserId, sale2ItemIds)

		setup.WithTransaction(t, func(transaction *queries.TransactionalDatabaseQuerier) {
			err := queries.RemoveSale(transaction, sale1.SaleID)
			require.NoError(t, err)
		})

		sale1Exists, err := queries.SaleWithIdExists(db, sale1.SaleID)
		require.NoError(t, err)
		require.False(t, sale1Exists)

		sale2Exists, err := queries.SaleWithIdExists(db, sale2.SaleID)
		require.NoError(t, err)
		require.True(t, sale2Exists)
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent sale", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			nonexistentSaleId := models.Id(999)
			setup.RequireNoSuchSales(t, nonexistentSaleId)

			setup.WithTransaction(t, func(transaction *queries.TransactionalDatabaseQuerier) {
				err := queries.RemoveSale(transaction, nonexistentSaleId)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSale)
			})
		})

		t.Run("Nonexistent sale in transaction", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			item := setup.Item(seller.UserId, aux.WithHidden(false), aux.WithFrozen(false))
			sale := setup.Sale(cashier.UserId, []models.Id{item.ItemID})

			nonexistentSaleId := models.Id(999)
			setup.RequireNoSuchSales(t, nonexistentSaleId)

			transactionErr := setup.WithTransactionErr(t, func(transaction *queries.TransactionalDatabaseQuerier) error {
				err := queries.RemoveSale(transaction, sale.SaleID)
				require.NoError(t, err)

				return queries.RemoveSale(transaction, nonexistentSaleId)
			})

			requireDatabaseWrappedError(t, transactionErr, dberr.ErrNoSuchSale)

			saleExists, saleExistsErr := queries.SaleWithIdExists(db, sale.SaleID)
			require.NoError(t, saleExistsErr)
			require.True(t, saleExists)
		})
	})
}
