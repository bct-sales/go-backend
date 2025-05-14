//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetSalesWithItem(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for saleCount := range []int{0, 1, 10} {
			testLabel := fmt.Sprintf("Sale count: %d", saleCount)

			t.Run(testLabel, func(t *testing.T) {
				saleCount := 0
				setup, db := NewDatabaseFixture()
				defer setup.Close()

				seller := setup.Seller()
				cashier := setup.Cashier()

				item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))

				saleIds := make([]models.Id, saleCount)
				for index := range saleIds {
					saleIds[index] = setup.Sale(cashier.UserId, []models.Id{item.ItemId})
				}

				actualSaleIds, err := queries.GetSalesWithItem(db, item.ItemId)
				require.NoError(t, err)
				require.Equal(t, saleIds, actualSaleIds)
			})
		}

		t.Run("Ignores other sales without the item", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()

			item1 := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
			item2 := setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithHidden(false))

			setup.Sale(cashier.UserId, []models.Id{item1.ItemId})
			setup.Sale(cashier.UserId, []models.Id{item2.ItemId})

			expectedSaleIds := []models.Id{item1.ItemId}
			actualSaleIds, err := queries.GetSalesWithItem(db, item1.ItemId)
			require.NoError(t, err)
			require.Equal(t, expectedSaleIds, actualSaleIds)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Invalid item ID", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()

			item1 := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
			item2 := setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithHidden(false))
			invalidItemId := models.Id(1000)
			setup.RequireNoSuchItem(t, invalidItemId)

			setup.Sale(cashier.UserId, []models.Id{item1.ItemId})
			setup.Sale(cashier.UserId, []models.Id{item2.ItemId})

			_, err := queries.GetSalesWithItem(db, invalidItemId)
			var noSuchItemError *queries.NoSuchItemError
			require.ErrorAs(t, err, &noSuchItemError)
		})
	})
}
