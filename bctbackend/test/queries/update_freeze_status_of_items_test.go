//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestUpdateFreezeStatusOfItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		selections := [][]models.Id{
			{},
			{1},
			{2},
			{1, 2},
			{1, 2, 3},
			{1, 2, 3, 4},
			{1, 2, 3, 5},
		}
		for _, selection := range selections {
			testLabel := fmt.Sprintf("Selection: %v", selection)
			t.Run(testLabel, func(t *testing.T) {
				setup, db := NewDatabaseFixture()
				defer setup.Close()

				seller := setup.Seller()

				itemIds := []models.Id{}
				for i := 0; i != 10; i++ {
					itemIds = append(itemIds, setup.Item(seller.UserId, aux.WithDummyData(i), aux.WithFrozen(false)).ItemId)
				}

				err := queries.UpdateFreezeStatusOfItems(db, selection, true)
				require.NoError(t, err)

				for _, itemId := range itemIds {
					isFrozen, err := queries.IsItemFrozen(db, itemId)
					expectedFrozen := slices.Contains(selection, itemId)
					assert.NoError(t, err)
					assert.Equal(t, expectedFrozen, isFrozen, "items[%d].frozen should be %v", itemId, expectedFrozen)
				}
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("No such item", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			err := queries.UpdateFreezeStatusOfItems(db, []models.Id{1}, true)
			require.Error(t, err)
		})

		t.Run("Hidden item", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			seller := setup.Seller()

			itemIds := []models.Id{}
			for i := 0; i != 10; i++ {
				itemIds = append(itemIds, setup.Item(seller.UserId, aux.WithDummyData(i), aux.WithFrozen(false)).ItemId)
			}
			itemIds = append(itemIds, setup.Item(seller.UserId, aux.WithDummyData(10), aux.WithFrozen(false), aux.WithHidden(true)).ItemId)

			err := queries.UpdateFreezeStatusOfItems(db, itemIds, true)
			var itemHiddenError *queries.ItemHiddenError
			require.ErrorAs(t, err, &itemHiddenError)

			for _, itemId := range itemIds {
				isFrozen, err := queries.IsItemFrozen(db, itemId)
				assert.NoError(t, err)
				assert.Equal(t, false, isFrozen, "item with id %d should not be frozen", itemId)
			}
		})
	})
}
