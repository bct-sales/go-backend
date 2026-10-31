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
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)
		cashier := setup.Cashier(t)
		sale1ItemIDs := []models.ID{
			setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithHidden(false)).ItemID,
			setup.Item(t, seller.UserID, aux.WithDummyData(2), aux.WithHidden(false)).ItemID,
		}
		sale2ItemIDs := []models.ID{
			setup.Item(t, seller.UserID, aux.WithDummyData(3), aux.WithHidden(false)).ItemID,
			setup.Item(t, seller.UserID, aux.WithDummyData(4), aux.WithHidden(false)).ItemID,
		}

		sale1 := setup.Sale(t, cashier.UserID, sale1ItemIDs)
		sale2 := setup.Sale(t, cashier.UserID, sale2ItemIDs)

		setup.WithTransaction(t, func(transaction *queries.TransactionalDatabaseQuerier) {
			err := queries.RemoveSale(t.Context(), transaction, sale1.SaleID)
			require.NoError(t, err)
		})

		sale1Exists, err := queries.SaleWithIDExists(t.Context(), db, sale1.SaleID)
		require.NoError(t, err)
		require.False(t, sale1Exists)

		sale2Exists, err := queries.SaleWithIDExists(t.Context(), db, sale2.SaleID)
		require.NoError(t, err)
		require.True(t, sale2Exists)
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent sale", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			nonexistentSaleID := setup.GenerateNonexistentSaleID(t)

			setup.WithTransaction(t, func(transaction *queries.TransactionalDatabaseQuerier) {
				err := queries.RemoveSale(t.Context(), transaction, nonexistentSaleID)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSale)
			})
		})

		t.Run("Nonexistent sale in transaction", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)
			cashier := setup.Cashier(t)
			item := setup.Item(t, seller.UserID, aux.WithHidden(false), aux.WithFrozen(false))
			sale := setup.Sale(t, cashier.UserID, []models.ID{item.ItemID})
			nonexistentSaleID := setup.GenerateNonexistentSaleID(t)

			transactionErr := setup.WithTransactionErr(t, func(transaction *queries.TransactionalDatabaseQuerier) error {
				err := queries.RemoveSale(t.Context(), transaction, sale.SaleID)
				require.NoError(t, err)

				return queries.RemoveSale(t.Context(), transaction, nonexistentSaleID)
			})

			requireDatabaseWrappedError(t, transactionErr, dberr.ErrNoSuchSale)

			saleExists, saleExistsErr := queries.SaleWithIDExists(t.Context(), db, sale.SaleID)
			require.NoError(t, saleExistsErr)
			require.True(t, saleExists)
		})
	})
}
