//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/test"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestGetItemWithId(t *testing.T) {
	t.Run("Nonexisting item", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		itemId := models.NewId(1)
		_, err := queries.GetItemWithId(db, itemId)

		var itemNotFoundError *queries.ItemNotFoundError
		if !assert.ErrorAs(t, err, &itemNotFoundError) {
			return
		}
	})

	t.Run("Existing item", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId
		itemId := test.AddItemToDatabase(db, sellerId, 1).ItemId

		item, err := queries.GetItemWithId(db, itemId)

		if !assert.NoError(t, err) {
			return
		}

		if !assert.Equal(t, itemId, item.ItemId) {
			return
		}
	})
}
