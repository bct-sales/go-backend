//go:build test

package queries

import (
	"bctbackend/algorithms"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestPartitionItemsByFrozenStatus(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("All visible, no nonexistent items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			frozenItems := setup.Items(seller.UserId, 10, aux.WithFrozen(true), aux.WithHidden(false))
			unfrozenItems := setup.Items(seller.UserId, 5, aux.WithFrozen(false), aux.WithHidden(false))
			allItems := append(
				algorithms.Map(frozenItems, func(i *models.Item) models.Id { return i.ItemId }),
				algorithms.Map(unfrozenItems, func(i *models.Item) models.Id { return i.ItemId })...)

			actualUnfrozen, actualFrozen, err := queries.PartitionItemsByFrozenStatus(db, allItems)
			require.NoError(t, err)
			require.Equal(t, len(frozenItems), actualFrozen.Len())
			require.Equal(t, len(unfrozenItems), actualUnfrozen.Len())
		})
	})
}
