//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Get only visible items", func(t *testing.T) {
			t.Run("No hidden items", func(t *testing.T) {
				for _, itemCount := range []int{0, 1, 2, 10} {
					testLabel := fmt.Sprintf("Item count: %d", itemCount)
					t.Run(testLabel, func(t *testing.T) {
						t.Parallel()

						setup, db := NewDatabaseFixture(WithDefaultCategories)
						defer setup.Close()

						seller := setup.Seller()
						items := setup.Items(seller.UserId, itemCount, aux.WithHidden(false))

						actualItems := []*models.Item{}
						err := queries.GetItems(db, queries.CollectTo(&actualItems), queries.OnlyVisibleItems, queries.AllRows())
						require.NoError(t, err)
						require.Equal(t, itemCount, len(actualItems))

						for i, item := range items {
							require.Equal(t, item, actualItems[i])
						}
					})
				}
			})

			t.Run("With hidden items", func(t *testing.T) {
				for _, itemCount := range []int{0, 1, 2, 10} {
					testLabel := fmt.Sprintf("Item count: %d", itemCount)
					t.Run(testLabel, func(t *testing.T) {
						setup, db := NewDatabaseFixture(WithDefaultCategories)
						defer setup.Close()

						seller := setup.Seller()
						setup.Items(seller.UserId, itemCount, aux.WithFrozen(false), aux.WithHidden(true))

						actualItems := []*models.Item{}
						err := queries.GetItems(db, queries.CollectTo(&actualItems), queries.OnlyVisibleItems, queries.AllRows())
						require.NoError(t, err)
						require.Equal(t, 0, len(actualItems))
					})
				}
			})
		})

		t.Run("Get all items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
			items = slices.Concat(items, setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(true)))

			actualItems := []*models.Item{}
			err := queries.GetItems(db, queries.CollectTo(&actualItems), queries.AllItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 20, len(actualItems))

			for i, item := range items {
				require.Equal(t, item, actualItems[i])
			}
		})

		t.Run("Get only hidden items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			setup.Items(seller.UserId, 10, aux.WithHidden(false))
			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(true))

			actualItems := []*models.Item{}
			err := queries.GetItems(db, queries.CollectTo(&actualItems), queries.OnlyHiddenItems, queries.AllRows())
			require.NoError(t, err)
			require.Equal(t, 10, len(actualItems))

			for i, item := range items {
				require.Equal(t, item, actualItems[i])
			}
		})

		t.Run("Get items 10-15", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			offset := 10
			limit := 5

			seller := setup.Seller()
			items := setup.Items(seller.UserId, 20, aux.WithHidden(false))

			actualItems := []*models.Item{}
			err := queries.GetItems(db, queries.CollectTo(&actualItems), queries.AllItems, queries.RowSelection(&offset, &limit))
			require.NoError(t, err)
			require.Equal(t, limit, len(actualItems))

			for index, actualItem := range actualItems {
				require.Equal(t, items[index+offset], actualItem)
			}
		})
	})
}
