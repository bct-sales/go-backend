//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetItemIds(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		items := setup.Items(seller.UserId, 10, aux.WithHidden(false))

		itemIds, err := queries.GetItemIds(db)
		require.NoError(t, err)
		require.ElementsMatch(t, itemIds, models.CollectItemIds(items))
	})
}
