//go:build test

package queries

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetItemsWithIDs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		selections := [][]models.ID{
			{},
			{1},
			{2},
			{1, 2},
			{2, 1},
			{1, 2, 3},
			{1, 2, 3, 4},
			{1, 2, 4},
			{1, 1},
			{1, 2, 3, 1, 2, 3},
		}
		for _, selection := range selections {
			testLabel := fmt.Sprintf("Selection: %v", selection)
			t.Run(testLabel, func(t *testing.T) {
				setup, db := NewDatabaseFixture(t, WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller(t)

				for i := 0; i != 10; i++ {
					setup.Item(t, seller.UserID, aux.WithDummyData(i), aux.WithHidden(false))
				}

				actual, err := queries.GetItemsWithIDs(t.Context(), db, selection)
				require.NoError(t, err)

				require.Equal(t, algorithms.NewSet(selection...).Len(), len(actual))
				for _, itemID := range selection {
					item, ok := actual[itemID]
					require.True(t, ok, "Item not found in result")
					require.Equal(t, itemID, item.ItemID)
				}
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent item", func(t *testing.T) {
			selections := [][]models.ID{
				{100},
				{1, 2, 3, 15},
			}
			for _, selection := range selections {
				testLabel := fmt.Sprintf("Selection: %v", selection)
				t.Run(testLabel, func(t *testing.T) {
					setup, db := NewDatabaseFixture(t, WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller(t)

					for i := 0; i != 10; i++ {
						setup.Item(t, seller.UserID, aux.WithDummyData(i), aux.WithHidden(false))
					}

					actual, err := queries.GetItemsWithIDs(t.Context(), db, selection)
					require.Nil(t, actual)
					requireDatabaseWrappedError(t, err, dberr.ErrNoSuchItem)
				})
			}
		})
	})
}
