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

func TestGetSellerFrozenItemCount(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Single seller", func(t *testing.T) {
			for _, frozenItemCount := range []int64{0, 1, 2, 10, 100} {
				for _, unfrozenItemCount := range []int64{0, 1, 2, 10, 100} {
					testLabel := fmt.Sprintf("Seller with %d frozen items and %d unfrozen items", frozenItemCount, unfrozenItemCount)
					t.Run(testLabel, func(t *testing.T) {
						setup, db := NewDatabaseFixture()
						defer setup.Close()

						seller := setup.Seller()

						for i := int64(0); i < frozenItemCount; i++ {
							setup.Item(seller.UserId, aux.WithDummyData(int(i)), aux.WithFrozen(true))
						}
						for i := int64(0); i < unfrozenItemCount; i++ {
							setup.Item(seller.UserId, aux.WithDummyData(int(i)), aux.WithFrozen(false))
						}

						actual, err := queries.GetSellerFrozenItemCount(db, seller.UserId)
						require.NoError(t, err)
						require.Equal(t, frozenItemCount, actual)
					})
				}
			}
		})
	})
}
