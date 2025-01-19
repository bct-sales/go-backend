//go:build test

package queries

import (
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/test"
	"bctbackend/test/setup"
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
				db := setup.OpenInitializedDatabase()
				defer db.Close()

				seller := test.AddSellerToDatabase(db)
				cashier := test.AddCashierToDatabase(db)
				item := test.AddItemToDatabase(db, seller.UserId, test.WithDummyData(1))

				for i := 0; i < sellCount; i++ {
					test.AddSaleToDatabase(db, cashier.UserId, []models.Id{item.ItemId})
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
		db := setup.OpenInitializedDatabase()
		defer db.Close()

		nonexistentItemId := models.Id(1)

		itemExists, err := queries.ItemWithIdExists(db, nonexistentItemId)
		require.NoError(t, err)
		require.False(t, itemExists)

		_, err = queries.GetSaleItemInformation(db, 1)
		require.Error(t, err)
	})
}
