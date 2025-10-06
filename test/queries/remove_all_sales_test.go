//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRemoveAllSales(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller1 := setup.Seller()
		seller2 := setup.Seller()
		seller3 := setup.Seller()
		cashier1 := setup.Cashier()
		cashier2 := setup.Cashier()
		items1 := setup.Items(seller1.UserID, 10, aux.WithHidden(false))
		items2 := setup.Items(seller2.UserID, 20, aux.WithHidden(false))
		items3 := setup.Items(seller3.UserID, 30, aux.WithHidden(false))
		sale1 := setup.Sale(cashier1.UserID, models.CollectItemIDs(items1))
		sale2 := setup.Sale(cashier1.UserID, models.CollectItemIDs(items2))
		sale3 := setup.Sale(cashier2.UserID, models.CollectItemIDs(items3))

		setup.WithTransaction(t, func(transaction *queries.TransactionalDatabaseQuerier) {
			err := queries.RemoveAllSales(transaction)
			require.NoError(t, err)
		})

		for _, sale := range []*models.Sale{sale1, sale2, sale3} {
			exists, err := queries.SaleWithIdExists(db, sale.SaleID)
			require.NoError(t, err)
			require.False(t, exists, "Sale should not exist after removal")
		}

		for _, item := range slices.Concat(items1, items2, items3) {
			exists, err := queries.ItemWithIDExists(db, item.ItemID)
			require.NoError(t, err)
			require.True(t, exists, "Item should still exist after sales removal")
		}
	})
}
