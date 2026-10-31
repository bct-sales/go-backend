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

func TestGetItemStatisticsQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Only visible items in count", func(t *testing.T) {
			t.Run("No hidden items", func(t *testing.T) {
				for _, count := range []int{0, 1, 2, 5, 10, 23} {
					testLabel := fmt.Sprintf("%d unhidden items", count)
					t.Run(testLabel, func(t *testing.T) {
						setup, db := NewDatabaseFixture(t, WithDefaultCategories)
						defer setup.Close()

						seller := setup.Seller(t)
						items := setup.Items(t, seller.UserID, count, aux.WithHidden(false))

						query := queries.NewGetItemStatisticsQuery()
						query.WithHidden(false)
						actual, err := query.Execute(t.Context(), db)
						require.NoError(t, err)
						require.Equal(t, count, actual.ItemCount)
						require.Equal(t, aux.ItemsTotalWorth(items), actual.TotalValueInCents)
					})
				}
			})

			t.Run("With hidden items", func(t *testing.T) {
				setup, db := NewDatabaseFixture(t, WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller(t)
				setup.Item(t, seller.UserID, aux.WithHidden(true))

				query := queries.NewGetItemStatisticsQuery()
				query.WithHidden(false)
				actual, err := query.Execute(t.Context(), db)
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
						setup, db := NewDatabaseFixture(t, WithDefaultCategories)
						defer setup.Close()

						seller := setup.Seller(t)
						items := setup.Items(t, seller.UserID, count, aux.WithHidden(false))

						query := queries.NewGetItemStatisticsQuery()
						actual, err := query.Execute(t.Context(), db)
						require.NoError(t, err)
						require.Equal(t, count, actual.ItemCount)
						require.Equal(t, aux.ItemsTotalWorth(items), actual.TotalValueInCents)
					})
				}
			})

			t.Run("With hidden items", func(t *testing.T) {
				setup, db := NewDatabaseFixture(t, WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller(t)
				item := setup.Item(t, seller.UserID, aux.WithFrozen(false), aux.WithHidden(true))

				query := queries.NewGetItemStatisticsQuery()
				actual, err := query.Execute(t.Context(), db)
				require.NoError(t, err)
				require.Equal(t, 1, actual.ItemCount)
				require.Equal(t, aux.ItemsTotalWorth([]*models.Item{item}), actual.TotalValueInCents)
			})
		})

		t.Run("Only hidden items in count", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)
			setup.Items(t, seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(false))
			hiddenItems := setup.Items(t, seller.UserID, 12, aux.WithFrozen(false), aux.WithHidden(true))

			query := queries.NewGetItemStatisticsQuery()
			query.WithHidden(true)
			actual, err := query.Execute(t.Context(), db)
			require.NoError(t, err)
			require.Equal(t, len(hiddenItems), actual.ItemCount)
			require.Equal(t, aux.ItemsTotalWorth(hiddenItems), actual.TotalValueInCents)
		})

		t.Run("With description filter", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)
			setup.Item(t, seller.UserID, aux.WithHidden(false), aux.WithDescription("abc"), aux.WithPriceInCents(50))
			setup.Item(t, seller.UserID, aux.WithHidden(false), aux.WithDescription("bcd"), aux.WithPriceInCents(100))
			setup.Item(t, seller.UserID, aux.WithHidden(false), aux.WithDescription("cde"), aux.WithPriceInCents(200))
			setup.Item(t, seller.UserID, aux.WithHidden(false), aux.WithDescription("edf"), aux.WithPriceInCents(400))

			query := queries.NewGetItemStatisticsQuery()
			query.WithDescription("d")
			actual, err := query.Execute(t.Context(), db)
			require.NoError(t, err)
			require.Equal(t, 3, actual.ItemCount)
			require.Equal(t, models.MoneyInCents(100+200+400), actual.TotalValueInCents)
		})

		t.Run("With category filter", func(t *testing.T) {
			t.Run("Two matching items", func(t *testing.T) {
				setup, db := NewDatabaseFixture(t)
				defer setup.Close()

				setup.Category(t, 1, "a")
				setup.Category(t, 2, "b")
				setup.Category(t, 3, "c")
				setup.Category(t, 4, "d")

				seller := setup.Seller(t)
				setup.Item(t, seller.UserID, aux.WithHidden(false), aux.WithItemCategory(1), aux.WithPriceInCents(50))
				setup.Item(t, seller.UserID, aux.WithHidden(false), aux.WithItemCategory(2), aux.WithPriceInCents(100))
				setup.Item(t, seller.UserID, aux.WithHidden(false), aux.WithItemCategory(2), aux.WithPriceInCents(200))
				setup.Item(t, seller.UserID, aux.WithHidden(false), aux.WithItemCategory(3), aux.WithPriceInCents(400))

				query := queries.NewGetItemStatisticsQuery()
				query.WithCategory(2)
				actual, err := query.Execute(t.Context(), db)
				require.NoError(t, err)
				require.Equal(t, 2, actual.ItemCount)
				require.Equal(t, models.MoneyInCents(100+200), actual.TotalValueInCents)
			})

			t.Run("Four matching items", func(t *testing.T) {
				setup, db := NewDatabaseFixture(t)
				defer setup.Close()

				setup.Category(t, 1, "a")
				setup.Category(t, 2, "b")
				setup.Category(t, 3, "c")

				seller := setup.Seller(t)
				setup.Items(t, seller.UserID, 1, aux.WithHidden(false), aux.WithItemCategory(1), aux.WithPriceInCents(50))
				setup.Items(t, seller.UserID, 4, aux.WithHidden(false), aux.WithItemCategory(2), aux.WithPriceInCents(100))
				setup.Items(t, seller.UserID, 3, aux.WithHidden(false), aux.WithItemCategory(3), aux.WithPriceInCents(200))

				query := queries.NewGetItemStatisticsQuery()
				query.WithCategory(2)
				actual, err := query.Execute(t.Context(), db)
				require.NoError(t, err)
				require.Equal(t, 4, actual.ItemCount)
				require.Equal(t, models.MoneyInCents(4*100), actual.TotalValueInCents)
			})
		})
	})
}
