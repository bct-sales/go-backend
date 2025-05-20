//go:build test

package queries

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetSaleItemInformation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, sellCount := range []int{0, 1, 2, 3, 10} {
			label := fmt.Sprintf("Sell count = %d", sellCount)

			t.Run(label, func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller()
				cashier := setup.Cashier()
				item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))

				for i := 0; i < sellCount; i++ {
					setup.Sale(cashier.UserId, []models.Id{item.ItemId})
				}

				itemInformation, err := queries.GetSaleItemInformation(db, item.ItemId)
				require.NoError(t, err)
				require.Equal(t, item.SellerId, itemInformation.SellerId)
				require.Equal(t, item.Description, itemInformation.Description)
				require.Equal(t, item.PriceInCents, itemInformation.PriceInCents)
				require.Equal(t, item.CategoryId, itemInformation.ItemCategoryId)
				require.Equal(t, itemInformation.SellCount, int64(sellCount))
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			nonexistentItemId := models.Id(1000)
			setup.RequireNoSuchItem(t, nonexistentItemId)

			_, err := queries.GetSaleItemInformation(db, 1)
			require.Error(t, err)
		})
	})
}
