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

					results := []*queries.UserWithItemCount{}
					err := queries.GetUsersWithItemCount(db, queries.CollectTo(&results))
					require.NoError(t, err)
					require.Len(t, results, 1)
					expected := &queries.UserWithItemCount{
						User:      *seller,
						ItemCount: int64(itemCount),
					}
					require.Equal(t, expected, results[0])
				})
			}
		})
	})
}
