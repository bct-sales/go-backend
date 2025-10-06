//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateFreezeStatusOfItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		selections := [][]models.ID{
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

				itemIds := []models.ID{}
				for i := 0; i != 10; i++ {
					itemIds = append(itemIds, setup.Item(seller.UserID, aux.WithDummyData(i), aux.WithFrozen(false), aux.WithHidden(false)).ItemID)
				}

				setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
					err := queries.UpdateFreezeStatusOfItems(db, selection, true)
					require.NoError(t, err)
				})

				for _, itemId := range itemIds {
					isFrozen, err := queries.IsItemFrozen(db, itemId)
					expectedFrozen := slices.Contains(selection, itemId)
					assert.NoError(t, err)
					assert.Equal(t, expectedFrozen, isFrozen, "item [%d] should have frozen=%v", itemId, expectedFrozen)
				}
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("No such item", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.UpdateFreezeStatusOfItems(db, []models.ID{1}, true)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchItem)
			})
		})

		t.Run("Cannot freeze hidden item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()

			itemIds := []models.ID{}
			for i := 0; i != 10; i++ {
				itemIds = append(itemIds, setup.Item(seller.UserID, aux.WithDummyData(i), aux.WithFrozen(false), aux.WithHidden(false)).ItemID)
			}
			itemIds = append(itemIds, setup.Item(seller.UserID, aux.WithDummyData(10), aux.WithFrozen(false), aux.WithHidden(true)).ItemID)

			setup.WithTransaction(t, func(db *queries.TransactionalDatabaseQuerier) {
				err := queries.UpdateFreezeStatusOfItems(db, itemIds, true)
				requireDatabaseWrappedError(t, err, dberr.ErrItemHidden)
			})

			for _, itemId := range itemIds {
				isFrozen, err := queries.IsItemFrozen(db, itemId)
				assert.NoError(t, err)
				assert.Equal(t, false, isFrozen, "item with id %d should not be frozen", itemId)
			}
		})
	})
}
