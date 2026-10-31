//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	models "bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSaleItemInformation(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, sellCount := range []int{0, 1, 2, 3, 10} {
			for i := range 10 {
				label := fmt.Sprintf("Sell count = %d, i = %d", sellCount, i)

				t.Run(label, func(t *testing.T) {
					setup, db := NewDatabaseFixture(t, WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller(t)
					cashier := setup.Cashier(t)
					item := setup.Item(t, seller.UserID, aux.WithDummyData(i), aux.WithHidden(false))

					for i := 0; i < sellCount; i++ {
						setup.Sale(t, cashier.UserID, []models.ID{item.ItemID})
					}

					itemInformation, err := queries.GetSaleItemInformation(t.Context(), db, item.ItemID)
					require.NoError(t, err)
					require.Equal(t, item.SellerID, itemInformation.SellerID)
					require.Equal(t, item.Description, itemInformation.Description)
					require.Equal(t, item.PriceInCents, itemInformation.PriceInCents)
					require.Equal(t, item.CategoryID, itemInformation.ItemCategoryID)
					require.Equal(t, item.Large, itemInformation.Large)
					require.Equal(t, itemInformation.SellCount, int64(sellCount))
				})
			}
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			nonexistentItemID := setup.GenerateNonexistentItemID(t)

			_, err := queries.GetSaleItemInformation(t.Context(), db, nonexistentItemID)
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchItem)
		})
	})
}
