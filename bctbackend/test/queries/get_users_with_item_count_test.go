//go:build test

package queries

import (
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetUsersWithItemCount(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Single seller", func(t *testing.T) {
			itemCounts := []int{0, 1, 2, 5, 12}

			for _, itemCount := range itemCounts {
				testLabel := fmt.Sprintf("ItemCount: %d", itemCount)
				t.Run(testLabel, func(t *testing.T) {
					t.Parallel()

					setup, db := NewDatabaseFixture(WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller()
					setup.Items(seller.UserId, itemCount, aux.WithHidden(false))

					actual := []*queries.UserWithItemCount{}
					err := queries.GetUsersWithItemCount(db, queries.AllItems, queries.CollectTo(&actual))
					require.NoError(t, err)
					require.Len(t, actual, 1)
					expected := &queries.UserWithItemCount{
						User:      *seller,
						ItemCount: itemCount,
					}
					require.Equal(t, expected, actual[0])
				})
			}
		})

		t.Run("Multiple sellers", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller1 := setup.Seller()
			seller2 := setup.Seller()
			seller3 := setup.Seller()
			setup.Items(seller1.UserId, 0, aux.WithHidden(false))
			setup.Items(seller2.UserId, 5, aux.WithHidden(false))
			setup.Items(seller3.UserId, 15, aux.WithHidden(false))

			actual := []*queries.UserWithItemCount{}
			err := queries.GetUsersWithItemCount(db, queries.AllItems, queries.CollectTo(&actual))
			require.NoError(t, err)
			require.Len(t, actual, 3)
			expected := []*queries.UserWithItemCount{
				{
					User:      *seller1,
					ItemCount: 0,
				},
				{
					User:      *seller2,
					ItemCount: 5,
				},
				{
					User:      *seller3,
					ItemCount: 15,
				},
			}
			require.Equal(t, expected, actual)
		})

		t.Run("Hidden items", func(t *testing.T) {
			t.Run("Exclude hidden items", func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				seller1 := setup.Seller()
				seller2 := setup.Seller()
				seller3 := setup.Seller()
				setup.Items(seller1.UserId, 0, aux.WithFrozen(false), aux.WithHidden(false))
				setup.Items(seller1.UserId, 3, aux.WithFrozen(false), aux.WithHidden(true))
				setup.Items(seller2.UserId, 5, aux.WithFrozen(false), aux.WithHidden(false))
				setup.Items(seller2.UserId, 5, aux.WithFrozen(false), aux.WithHidden(true))
				setup.Items(seller3.UserId, 15, aux.WithFrozen(false), aux.WithHidden(false))
				setup.Items(seller3.UserId, 4, aux.WithFrozen(false), aux.WithHidden(true))

				actual := []*queries.UserWithItemCount{}
				err := queries.GetUsersWithItemCount(db, queries.OnlyVisibleItems, queries.CollectTo(&actual))
				require.NoError(t, err)
				require.Len(t, actual, 3)
				expected := []*queries.UserWithItemCount{
					{
						User:      *seller1,
						ItemCount: 0,
					},
					{
						User:      *seller2,
						ItemCount: 5,
					},
					{
						User:      *seller3,
						ItemCount: 15,
					},
				}
				require.Equal(t, expected, actual)
			})

			t.Run("Include hidden items", func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				seller1 := setup.Seller()
				seller2 := setup.Seller()
				seller3 := setup.Seller()
				setup.Items(seller1.UserId, 0, aux.WithFrozen(false), aux.WithHidden(false))
				setup.Items(seller1.UserId, 3, aux.WithFrozen(false), aux.WithHidden(true))
				setup.Items(seller2.UserId, 5, aux.WithFrozen(false), aux.WithHidden(false))
				setup.Items(seller2.UserId, 5, aux.WithFrozen(false), aux.WithHidden(true))
				setup.Items(seller3.UserId, 15, aux.WithFrozen(false), aux.WithHidden(false))
				setup.Items(seller3.UserId, 4, aux.WithFrozen(false), aux.WithHidden(true))

				actual := []*queries.UserWithItemCount{}
				err := queries.GetUsersWithItemCount(db, queries.AllItems, queries.CollectTo(&actual))
				require.NoError(t, err)
				require.Len(t, actual, 3)
				expected := []*queries.UserWithItemCount{
					{
						User:      *seller1,
						ItemCount: 3,
					},
					{
						User:      *seller2,
						ItemCount: 10,
					},
					{
						User:      *seller3,
						ItemCount: 19,
					},
				}
				require.Equal(t, expected, actual)
			})

			t.Run("Only hidden items", func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				seller1 := setup.Seller()
				seller2 := setup.Seller()
				seller3 := setup.Seller()
				setup.Items(seller1.UserId, 0, aux.WithFrozen(false), aux.WithHidden(false))
				setup.Items(seller1.UserId, 3, aux.WithFrozen(false), aux.WithHidden(true))
				setup.Items(seller2.UserId, 5, aux.WithFrozen(false), aux.WithHidden(false))
				setup.Items(seller2.UserId, 5, aux.WithFrozen(false), aux.WithHidden(true))
				setup.Items(seller3.UserId, 15, aux.WithFrozen(false), aux.WithHidden(false))
				setup.Items(seller3.UserId, 4, aux.WithFrozen(false), aux.WithHidden(true))

				actual := []*queries.UserWithItemCount{}
				err := queries.GetUsersWithItemCount(db, queries.OnlyHiddenItems, queries.CollectTo(&actual))
				require.NoError(t, err)
				require.Len(t, actual, 3)
				expected := []*queries.UserWithItemCount{
					{
						User:      *seller1,
						ItemCount: 3,
					},
					{
						User:      *seller2,
						ItemCount: 5,
					},
					{
						User:      *seller3,
						ItemCount: 4,
					},
				}
				require.Equal(t, expected, actual)
			})
		})
	})
}
