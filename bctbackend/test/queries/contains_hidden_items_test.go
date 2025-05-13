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

func TestContainsHiddenItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("No hidden items", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			seller := setup.Seller()
			items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
			itemIds := algorithms.Map(items, func(item *models.Item) models.Id { return item.ItemId })

			result, err := queries.ContainsHiddenItems(db, itemIds)
			require.NoError(t, err)
			require.False(t, result)
		})
	})
}
