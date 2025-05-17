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
					setup, db := NewDatabaseFixture()
					defer setup.Close()

					seller := setup.Seller()
					setup.Items(seller.UserId, itemCount, aux.WithHidden(false))

					actual := []*queries.UserWithItemCount{}
					err := queries.GetUsersWithItemCount(db, queries.CollectTo(&actual))
					require.NoError(t, err)
					require.Len(t, actual, 1)
					expected := &queries.UserWithItemCount{
						User:      *seller,
						ItemCount: int64(itemCount),
					}
					require.Equal(t, expected, actual[0])
				})
			}
		})

		t.Run("Multiple sellers", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			seller1 := setup.Seller()
			seller2 := setup.Seller()
			seller3 := setup.Seller()
			setup.Items(seller1.UserId, 0, aux.WithHidden(false))
			setup.Items(seller2.UserId, 5, aux.WithHidden(false))
			setup.Items(seller3.UserId, 15, aux.WithHidden(false))

			actual := []*queries.UserWithItemCount{}
			err := queries.GetUsersWithItemCount(db, queries.CollectTo(&actual))
			require.NoError(t, err)
			require.Len(t, actual, 3)
			expected := []*queries.UserWithItemCount{
				{
					User:      *seller1,
					ItemCount: int64(0),
				},
				{
					User:      *seller2,
					ItemCount: int64(5),
				},
				{
					User:      *seller3,
					ItemCount: int64(15),
				},
			}
			require.Equal(t, expected, actual)
		})
	})
}
