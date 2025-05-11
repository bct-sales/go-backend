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

func TestGetItemsWithIds(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		selections := [][]models.Id{
			[]models.Id{},
			[]models.Id{1},
			[]models.Id{2},
			[]models.Id{1, 2},
			[]models.Id{2, 1},
			[]models.Id{1, 2, 3},
			[]models.Id{1, 2, 3, 4},
			[]models.Id{1, 2, 4},
		}
		for _, selection := range selections {
			testLabel := fmt.Sprintf("Selection: %v", selection)
			t.Run(testLabel, func(t *testing.T) {
				setup, db := NewDatabaseFixture()
				defer setup.Close()

				seller := setup.Seller()

				for i := 0; i != 10; i++ {
					setup.Item(seller.UserId, aux.WithDummyData(i))
				}

				actual, err := queries.GetItemsWithIds(db, selection)
				require.NoError(t, err)

				require.Len(t, actual, len(selection))
				for _, itemId := range selection {
					item, ok := actual[itemId]
					require.True(t, ok, "Item not found in result")
					require.Equal(t, itemId, item.ItemId)
				}
			})
		}
	})
}
