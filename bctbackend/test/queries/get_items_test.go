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

func TestGetItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Get only visible items", func(t *testing.T) {
			t.Run("No hidden items", func(t *testing.T) {
				for _, itemCount := range []int{0, 1, 2, 10} {
					testLabel := fmt.Sprintf("Item count: %d", itemCount)
					t.Run(testLabel, func(t *testing.T) {
						setup, db := NewDatabaseFixture()
						defer setup.Close()

						seller := setup.Seller()
						items := setup.Items(seller.UserId, itemCount, aux.WithHidden(false))

						actualItems := []*models.Item{}
						err := queries.GetItems(db, queries.CollectTo(&actualItems), queries.OnlyVisibleItems)
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
						setup, db := NewDatabaseFixture()
						defer setup.Close()

						seller := setup.Seller()
						setup.Items(seller.UserId, itemCount, aux.WithHidden(true))

						actualItems := []*models.Item{}
						err := queries.GetItems(db, queries.CollectTo(&actualItems), queries.OnlyVisibleItems)
						require.NoError(t, err)
						require.Equal(t, 0, len(actualItems))
					})
				}
			})
		})

		t.Run("Get all items", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
			items = append(items, setup.Items(seller.UserId, 10, aux.WithHidden(true))...)

			actualItems := []*models.Item{}
			err := queries.GetItems(db, queries.CollectTo(&actualItems), queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, 20, len(actualItems))

			for i, item := range items {
				require.Equal(t, item, actualItems[i])
			}
		})
	})
}
