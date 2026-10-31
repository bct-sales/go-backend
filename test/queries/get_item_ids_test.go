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

func TestGetItemIDs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)
		items := setup.Items(t, seller.UserID, 10, aux.WithHidden(false))

		itemIDs, err := queries.GetItemIDs(t.Context(), db)
		require.NoError(t, err)
		require.ElementsMatch(t, itemIDs, models.CollectItemIDs(items))
	})
}
