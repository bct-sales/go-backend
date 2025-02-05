//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetItemWithId(t *testing.T) {
	t.Run("Nonexisting item", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		itemId := models.NewId(1)
		_, err := queries.GetItemWithId(db, itemId)
		var NoSuchItemError *queries.NoSuchItemError
		require.ErrorAs(t, err, &NoSuchItemError)
	})

	t.Run("Existing item", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId
		itemId := AddItemToDatabase(db, sellerId, WithDummyData(1)).ItemId

		item, err := queries.GetItemWithId(db, itemId)
		require.NoError(t, err)
		require.Equal(t, itemId, item.ItemId)
	})
}
