//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSalesWithItem(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for saleCount := range []int{0, 1, 10} {
			testLabel := fmt.Sprintf("Sale count: %d", saleCount)

			t.Run(testLabel, func(t *testing.T) {
				saleCount := 0
				setup, db := NewDatabaseFixture(t, WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller(t)
				cashier := setup.Cashier(t)

				item := setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))

				saleIDs := make([]models.ID, saleCount)
				for index := range saleIDs {
					saleIDs[index] = setup.Sale(t, cashier.UserID, []models.ID{item.ItemID}).SaleID
				}

				actualSaleIDs, err := queries.GetSalesWithItem(t.Context(), db, item.ItemID)
				require.NoError(t, err)
				require.Equal(t, saleIDs, actualSaleIDs)
			})
		}

		t.Run("Ignores other sales without the item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)
			cashier := setup.Cashier(t)

			item1 := setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))
			item2 := setup.Item(t, seller.UserID, aux.WithDummyData(2), aux.WithHidden(false))

			setup.Sale(t, cashier.UserID, []models.ID{item1.ItemID})
			setup.Sale(t, cashier.UserID, []models.ID{item2.ItemID})

			expectedSaleIDs := []models.ID{item1.ItemID}
			actualSaleIDs, err := queries.GetSalesWithItem(t.Context(), db, item1.ItemID)
			require.NoError(t, err)
			require.Equal(t, expectedSaleIDs, actualSaleIDs)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Invalid item ID", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)
			cashier := setup.Cashier(t)

			item1 := setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))
			item2 := setup.Item(t, seller.UserID, aux.WithDummyData(2), aux.WithHidden(false))
			invalidItemID := setup.GenerateNonexistentItemID(t)

			setup.Sale(t, cashier.UserID, []models.ID{item1.ItemID})
			setup.Sale(t, cashier.UserID, []models.ID{item2.ItemID})

			_, err := queries.GetSalesWithItem(t.Context(), db, invalidItemID)
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchItem)
		})
	})
}
