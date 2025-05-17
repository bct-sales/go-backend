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

func TestCheckItemsExistence(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("All items exist", func(t *testing.T) {
			selections := [][]models.Id{
				{},
				{1},
				{2},
				{1, 2},
				{1, 2, 3},
				{1, 2, 3, 4},
				{1, 2, 3, 5},
				{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
				{1, 1},
			}
			for _, selection := range selections {
				testLabel := fmt.Sprintf("Selection: %v", selection)
				t.Run(testLabel, func(t *testing.T) {
					setup, db := NewDatabaseFixture()
					defer setup.Close()

					seller := setup.Seller()

					itemIds := []models.Id{}
					for i := 0; i != 10; i++ {
						itemIds = append(itemIds, setup.Item(seller.UserId, aux.WithDummyData(i), aux.WithFrozen(false), aux.WithHidden(false)).ItemId)
					}

					actual, err := queries.ItemsExist(db, selection)
					require.NoError(t, err)
					require.True(t, actual)
				})
			}
		})

		t.Run("Not all items exist", func(t *testing.T) {
			selections := [][]models.Id{
				{11},
				{1, 11},
			}
			for _, selection := range selections {
				testLabel := fmt.Sprintf("Selection: %v", selection)
				t.Run(testLabel, func(t *testing.T) {
					setup, db := NewDatabaseFixture()
					defer setup.Close()

					seller := setup.Seller()

					itemIds := []models.Id{}
					for i := 0; i != 10; i++ {
						itemIds = append(itemIds, setup.Item(seller.UserId, aux.WithDummyData(i), aux.WithFrozen(false), aux.WithHidden(false)).ItemId)
					}

					actual, err := queries.ItemsExist(db, selection)
					require.NoError(t, err)
					require.False(t, actual)
				})
			}
		})
	})
}
