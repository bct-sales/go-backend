//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/test"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetItemWithId(t *testing.T) {
	t.Run("Nonexisting item", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		itemId := models.NewId(1)
		_, err := queries.GetItemWithId(db, itemId)
		var itemNotFoundError *queries.ItemNotFoundError
		require.ErrorAs(t, err, &itemNotFoundError)
	})

	t.Run("Existing item", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId
		itemId := test.AddItemToDatabase(db, sellerId, test.WithDummyData(1)).ItemId

		item, err := queries.GetItemWithId(db, itemId)
		require.NoError(t, err)
		require.Equal(t, itemId, item.ItemId)
	})
}
