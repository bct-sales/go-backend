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
)

func TestCountItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Only visible items in count", func(t *testing.T) {
			t.Run("No hidden items", func(t *testing.T) {
				for _, count := range []int{0, 1, 2, 5, 10, 23} {
					testLabel := fmt.Sprintf("%d unhidden items", count)
					t.Run(testLabel, func(t *testing.T) {
						setup, db := NewDatabaseFixture(WithDefaultCategories)
						defer setup.Close()

						seller := setup.Seller()
						items := setup.Items(seller.UserId, count, aux.WithHidden(false))

						actual, err := queries.CountItems(db, queries.OnlyVisibleItems)
						require.NoError(t, err)
						require.Equal(t, count, actual.ItemCount)
						require.Equal(t, aux.ItemsTotalWorth(items), actual.TotalValueInCents)
					})
				}
			})

			t.Run("With hidden items", func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller()
				setup.Item(seller.UserId, aux.WithHidden(true))

				actual, err := queries.CountItems(db, queries.OnlyVisibleItems)
				require.NoError(t, err)
				require.Equal(t, 0, actual.ItemCount)
				require.Equal(t, models.MoneyInCents(0), actual.TotalValueInCents)
			})
		})

		t.Run("All items in count", func(t *testing.T) {
			t.Run("No hidden items", func(t *testing.T) {
				for _, count := range []int{0, 1, 2, 5, 10, 23} {
					testLabel := fmt.Sprintf("%d unhidden items", count)
					t.Run(testLabel, func(t *testing.T) {
						setup, db := NewDatabaseFixture(WithDefaultCategories)
						defer setup.Close()

						seller := setup.Seller()
						items := setup.Items(seller.UserId, count, aux.WithHidden(false))

						actual, err := queries.CountItems(db, queries.AllItems)
						require.NoError(t, err)
						require.Equal(t, count, actual.ItemCount)
						require.Equal(t, aux.ItemsTotalWorth(items), actual.TotalValueInCents)
					})
				}
			})

			t.Run("With hidden items", func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller()
				item := setup.Item(seller.UserId, aux.WithFrozen(false), aux.WithHidden(true))

				actual, err := queries.CountItems(db, queries.AllItems)
				require.NoError(t, err)
				require.Equal(t, 1, actual.ItemCount)
				require.Equal(t, aux.ItemsTotalWorth([]*models.Item{item}), actual.TotalValueInCents)
			})
		})

		t.Run("Only hidden items in count", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(false))
			hiddenItems := setup.Items(seller.UserId, 12, aux.WithFrozen(false), aux.WithHidden(true))

			actual, err := queries.CountItems(db, queries.OnlyHiddenItems)
			require.NoError(t, err)
			require.Equal(t, len(hiddenItems), actual.ItemCount)
			require.Equal(t, aux.ItemsTotalWorth(hiddenItems), actual.TotalValueInCents)
		})
	})
}
