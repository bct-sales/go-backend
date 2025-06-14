//go:build test

package queries

import (
	"bctbackend/algorithms"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetItemsWithSaleCounts(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("All items from all sellers", func(t *testing.T) {
			for _, itemCount := range []int{0, 1, 2, 5, 15} {
				testLabel := fmt.Sprintf("itemCount=%d", itemCount)

				t.Run(testLabel, func(t *testing.T) {
					setup, db := NewDatabaseFixture(WithDefaultCategories)
					defer setup.Close()

					seller1 := setup.Seller()
					seller2 := setup.Seller()
					seller3 := setup.Seller()
					itemsIds1 := setup.Items(seller1.UserId, itemCount, aux.WithHidden(false))
					itemsIds2 := setup.Items(seller2.UserId, itemCount, aux.WithHidden(false))
					itemsIds3 := setup.Items(seller3.UserId, itemCount, aux.WithHidden(false))
					expectedItems := slices.Concat(itemsIds1, itemsIds2, itemsIds3)

					actualItems, err := queries.GetItemsWithSaleCounts(db, queries.AllItems, nil)
					require.NoError(t, err)
					require.Equal(t, itemCount*3, len(actualItems))

					expectedItemIds := algorithms.Map(expectedItems, func(item *models.Item) models.Id { return item.ItemID })
					actualItemIds := algorithms.Map(actualItems, func(item *queries.ItemWithSaleCount) models.Id { return item.ItemID })
					require.ElementsMatch(t, expectedItemIds, actualItemIds)
				})
			}
		})
	})
}
