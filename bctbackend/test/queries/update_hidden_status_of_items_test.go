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

func TestUpdateHiddenStatusOfItems(t *testing.T) {
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
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller()

				itemIds := []models.Id{}
				for i := 0; i != 10; i++ {
					itemIds = append(itemIds, setup.Item(seller.UserId, aux.WithDummyData(i), aux.WithHidden(false), aux.WithFrozen(false)).ItemId)
				}

				err := queries.UpdateHiddenStatusOfItems(db, selection, true)
				require.NoError(t, err)

				for _, itemId := range itemIds {
					isHidden, err := queries.IsItemHidden(db, itemId)
					expectedHidden := slices.Contains(selection, itemId)
					assert.NoError(t, err)
					assert.Equal(t, expectedHidden, isHidden, "item %d should have hidden=%v", itemId, expectedHidden)
				}
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("No such item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			err := queries.UpdateHiddenStatusOfItems(db, []models.Id{1}, true)
			require.Error(t, err)
		})

		t.Run("Cannot hide frozen item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()

			itemIds := []models.Id{}
			for i := 0; i != 10; i++ {
				itemIds = append(itemIds, setup.Item(seller.UserId, aux.WithDummyData(i), aux.WithHidden(false), aux.WithFrozen(false)).ItemId)
			}
			itemIds = append(itemIds, setup.Item(seller.UserId, aux.WithDummyData(10), aux.WithHidden(false), aux.WithFrozen(true)).ItemId)

			err := queries.UpdateHiddenStatusOfItems(db, itemIds, true)
			require.ErrorIs(t, err, queries.ErrItemFrozen)

			for _, itemId := range itemIds {
				isHidden, err := queries.IsItemHidden(db, itemId)
				assert.NoError(t, err)
				assert.Equal(t, false, isHidden, "item with id %d should not be hidden", itemId)
			}
		})
	})
}
