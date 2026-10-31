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
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller1 := setup.Seller(t)
		seller2 := setup.Seller(t)
		seller3 := setup.Seller(t)
		cashier1 := setup.Cashier(t)
		cashier2 := setup.Cashier(t)
		items1 := setup.Items(t, seller1.UserID, 10, aux.WithHidden(false))
		items2 := setup.Items(t, seller2.UserID, 20, aux.WithHidden(false))
		items3 := setup.Items(t, seller3.UserID, 30, aux.WithHidden(false))
		sale1 := setup.Sale(t, cashier1.UserID, models.CollectItemIDs(items1))
		sale2 := setup.Sale(t, cashier1.UserID, models.CollectItemIDs(items2))
		sale3 := setup.Sale(t, cashier2.UserID, models.CollectItemIDs(items3))

		setup.WithTransaction(t, func(transaction *queries.TransactionalDatabaseQuerier) {
			err := queries.RemoveAllSales(t.Context(), transaction)
			require.NoError(t, err)
		})

		for _, sale := range []*models.Sale{sale1, sale2, sale3} {
			exists, err := queries.SaleWithIDExists(t.Context(), db, sale.SaleID)
			require.NoError(t, err)
			require.False(t, exists, "Sale should not exist after removal")
		}

		for _, item := range slices.Concat(items1, items2, items3) {
			exists, err := queries.ItemWithIDExists(t.Context(), db, item.ItemID)
			require.NoError(t, err)
			require.True(t, exists, "Item should still exist after sales removal")
		}
	})
}
