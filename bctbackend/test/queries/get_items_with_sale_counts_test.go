//go:build test

package queries

import (
	"bctbackend/algorithms"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetItemsWithSaleCounts(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("All items from all sellers, zero sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller1 := setup.Seller()
			seller2 := setup.Seller()
			seller3 := setup.Seller()
			visibleItems1 := setup.Items(seller1.UserId, 1, aux.WithHidden(false), aux.WithFrozen(false))
			hiddenItems1 := setup.Items(seller1.UserId, 2, aux.WithHidden(true), aux.WithFrozen(false))
			visibleItems2 := setup.Items(seller2.UserId, 4, aux.WithHidden(false), aux.WithFrozen(false))
			hiddenItems2 := setup.Items(seller2.UserId, 8, aux.WithHidden(true), aux.WithFrozen(false))
			visibleItems3 := setup.Items(seller3.UserId, 16, aux.WithHidden(false), aux.WithFrozen(false))
			hiddenItems3 := setup.Items(seller3.UserId, 32, aux.WithHidden(true), aux.WithFrozen(false))

			expectedItems := slices.Concat(visibleItems1, visibleItems2, visibleItems3, hiddenItems1, hiddenItems2, hiddenItems3)
			actualItems, err := queries.GetItemsWithSaleCounts(db, queries.AllItems, nil)
			require.NoError(t, err)
			require.Equal(t, len(expectedItems), len(actualItems))

			expectedItemIds := algorithms.Map(expectedItems, func(item *models.Item) models.Id { return item.ItemID })
			actualItemIds := algorithms.Map(actualItems, func(item *queries.ItemWithSaleCount) models.Id { return item.ItemID })
			require.ElementsMatch(t, expectedItemIds, actualItemIds)
			for _, item := range actualItems {
				require.Equal(t, 0, item.SaleCount) // No sales in the fixture
			}
		})

		t.Run("All items from seller 1, zero sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller1 := setup.Seller()
			seller2 := setup.Seller()
			seller3 := setup.Seller()
			visibleItems1 := setup.Items(seller1.UserId, 1, aux.WithHidden(false), aux.WithFrozen(false))
			hiddenItems1 := setup.Items(seller1.UserId, 2, aux.WithHidden(true), aux.WithFrozen(false))
			setup.Items(seller2.UserId, 4, aux.WithHidden(false), aux.WithFrozen(false))
			setup.Items(seller2.UserId, 8, aux.WithHidden(true), aux.WithFrozen(false))
			setup.Items(seller3.UserId, 16, aux.WithHidden(false), aux.WithFrozen(false))
			setup.Items(seller3.UserId, 32, aux.WithHidden(true), aux.WithFrozen(false))

			expectedItems := slices.Concat(visibleItems1, hiddenItems1)
			sellerId := models.Id(1)
			actualItems, err := queries.GetItemsWithSaleCounts(db, queries.AllItems, &sellerId)
			require.NoError(t, err)
			require.Equal(t, len(expectedItems), len(actualItems))

			expectedItemIds := algorithms.Map(expectedItems, func(item *models.Item) models.Id { return item.ItemID })
			actualItemIds := algorithms.Map(actualItems, func(item *queries.ItemWithSaleCount) models.Id { return item.ItemID })
			require.ElementsMatch(t, expectedItemIds, actualItemIds)
			for _, item := range actualItems {
				require.Equal(t, 0, item.SaleCount) // No sales in the fixture
			}
		})

		t.Run("All items from seller 2, zero sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller1 := setup.Seller()
			seller2 := setup.Seller()
			seller3 := setup.Seller()
			setup.Items(seller1.UserId, 1, aux.WithHidden(false), aux.WithFrozen(false))
			setup.Items(seller1.UserId, 2, aux.WithHidden(true), aux.WithFrozen(false))
			visibleItems2 := setup.Items(seller2.UserId, 4, aux.WithHidden(false), aux.WithFrozen(false))
			hiddenItems2 := setup.Items(seller2.UserId, 8, aux.WithHidden(true), aux.WithFrozen(false))
			setup.Items(seller3.UserId, 16, aux.WithHidden(false), aux.WithFrozen(false))
			setup.Items(seller3.UserId, 32, aux.WithHidden(true), aux.WithFrozen(false))

			expectedItems := slices.Concat(visibleItems2, hiddenItems2)
			sellerId := models.Id(2)
			actualItems, err := queries.GetItemsWithSaleCounts(db, queries.AllItems, &sellerId)
			require.NoError(t, err)
			require.Equal(t, len(expectedItems), len(actualItems))

			expectedItemIds := algorithms.Map(expectedItems, func(item *models.Item) models.Id { return item.ItemID })
			actualItemIds := algorithms.Map(actualItems, func(item *queries.ItemWithSaleCount) models.Id { return item.ItemID })
			require.ElementsMatch(t, expectedItemIds, actualItemIds)
			for _, item := range actualItems {
				require.Equal(t, 0, item.SaleCount) // No sales in the fixture
			}
		})

		t.Run("All visible items from all sellers, zero sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller1 := setup.Seller()
			seller2 := setup.Seller()
			seller3 := setup.Seller()
			visibleItems1 := setup.Items(seller1.UserId, 1, aux.WithHidden(false), aux.WithFrozen(false))
			setup.Items(seller1.UserId, 2, aux.WithHidden(true), aux.WithFrozen(false))
			visibleItems2 := setup.Items(seller2.UserId, 4, aux.WithHidden(false), aux.WithFrozen(false))
			setup.Items(seller2.UserId, 8, aux.WithHidden(true), aux.WithFrozen(false))
			visibleItems3 := setup.Items(seller3.UserId, 16, aux.WithHidden(false), aux.WithFrozen(false))
			setup.Items(seller3.UserId, 32, aux.WithHidden(true), aux.WithFrozen(false))

			expectedItems := slices.Concat(visibleItems1, visibleItems2, visibleItems3)
			actualItems, err := queries.GetItemsWithSaleCounts(db, queries.OnlyVisibleItems, nil)
			require.NoError(t, err)
			require.Equal(t, len(expectedItems), len(actualItems))

			expectedItemIds := algorithms.Map(expectedItems, func(item *models.Item) models.Id { return item.ItemID })
			actualItemIds := algorithms.Map(actualItems, func(item *queries.ItemWithSaleCount) models.Id { return item.ItemID })
			require.ElementsMatch(t, expectedItemIds, actualItemIds)
			for _, item := range actualItems {
				require.Equal(t, 0, item.SaleCount) // No sales in the fixture
			}
		})

		t.Run("All hidden items from all sellers, zero sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller1 := setup.Seller()
			seller2 := setup.Seller()
			seller3 := setup.Seller()
			setup.Items(seller1.UserId, 1, aux.WithHidden(false), aux.WithFrozen(false))
			hiddenItems1 := setup.Items(seller1.UserId, 2, aux.WithHidden(true), aux.WithFrozen(false))
			setup.Items(seller2.UserId, 4, aux.WithHidden(false), aux.WithFrozen(false))
			hiddenItems2 := setup.Items(seller2.UserId, 8, aux.WithHidden(true), aux.WithFrozen(false))
			setup.Items(seller3.UserId, 16, aux.WithHidden(false), aux.WithFrozen(false))
			hiddenItems3 := setup.Items(seller3.UserId, 32, aux.WithHidden(true), aux.WithFrozen(false))

			expectedItems := slices.Concat(hiddenItems1, hiddenItems2, hiddenItems3)
			actualItems, err := queries.GetItemsWithSaleCounts(db, queries.OnlyHiddenItems, nil)
			require.NoError(t, err)
			require.Equal(t, len(expectedItems), len(actualItems))

			expectedItemIds := algorithms.Map(expectedItems, func(item *models.Item) models.Id { return item.ItemID })
			actualItemIds := algorithms.Map(actualItems, func(item *queries.ItemWithSaleCount) models.Id { return item.ItemID })
			require.ElementsMatch(t, expectedItemIds, actualItemIds)
			for _, item := range actualItems {
				require.Equal(t, 0, item.SaleCount) // No sales in the fixture
			}
		})

		t.Run("With sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller1 := setup.Seller()
			seller2 := setup.Seller()
			seller3 := setup.Seller()
			items1 := setup.Items(seller1.UserId, 1, aux.WithHidden(false), aux.WithFrozen(false))
			items2 := setup.Items(seller2.UserId, 4, aux.WithHidden(false), aux.WithFrozen(false))
			items3 := setup.Items(seller3.UserId, 16, aux.WithHidden(false), aux.WithFrozen(false))
			allItems := slices.Concat(items1, items2, items3)

			cashier := setup.Cashier()
			for _, item := range allItems {
				for range item.ItemID.Int64() {
					setup.Sale(cashier.UserId, []models.Id{item.ItemID})
				}
			}

			actualItems, err := queries.GetItemsWithSaleCounts(db, queries.AllItems, nil)
			require.NoError(t, err)
			require.Equal(t, len(allItems), len(actualItems))

			expectedItemIds := algorithms.Map(allItems, func(item *models.Item) models.Id { return item.ItemID })
			actualItemIds := algorithms.Map(actualItems, func(item *queries.ItemWithSaleCount) models.Id { return item.ItemID })
			require.ElementsMatch(t, expectedItemIds, actualItemIds)
			for _, item := range actualItems {
				expectedSaleCount := item.ItemID.Int64()
				actualSaleCount := int64(item.SaleCount)
				require.Equal(t, expectedSaleCount, actualSaleCount, "Item %d has unexpected sale count: expected %d, got %d", item.ItemID, expectedSaleCount, actualSaleCount)
			}
		})
	})
}
