//go:build test

package queries

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test"
	aux "bctbackend/test/helpers"
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
				setup, db := Setup()
				defer setup.Close()

				seller := setup.Seller()
				cashier := setup.Cashier()
				item := setup.Item(seller.UserId, aux.WithDummyData(1))

				for i := 0; i < sellCount; i++ {
					setup.Sale(cashier.UserId, []models.Id{item.ItemId})
				}

				itemInformation, err := queries.GetSaleItemInformation(db, item.ItemId)
				require.NoError(t, err)
				require.Equal(t, item.Description, itemInformation.Description)
				require.Equal(t, item.PriceInCents, itemInformation.PriceInCents)
				require.Equal(t, item.CategoryId, itemInformation.ItemCategoryId)
				require.Equal(t, itemInformation.SellCount, int64(sellCount))
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		nonexistentItemId := models.Id(1)

		itemExists, err := queries.ItemWithIdExists(db, nonexistentItemId)
		require.NoError(t, err)
		require.False(t, itemExists)

		_, err = queries.GetSaleItemInformation(db, 1)
		require.Error(t, err)
	})
}
