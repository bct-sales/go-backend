//go:build test

package queries

import (
	"bctbackend/database/queries"
	"bctbackend/test"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestGetItems(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		items, err := queries.GetItems(db)
		if !assert.NoError(t, err) {
			return
		}

		if !assert.Equal(t, 0, len(items)) {
			return
		}
	})

	t.Run("One item", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId
		itemId := test.AddItemToDatabase(db, sellerId, 1).ItemId

		items, err := queries.GetItems(db)

		if !assert.NoError(t, err) {
			return
		}

		if !assert.Equal(t, 1, len(items)) {
			return
		}

		if !assert.Equal(t, itemId, items[0].ItemId) {
			return
		}
	})

	t.Run("Two items", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId
		item1Id := test.AddItemToDatabase(db, sellerId, 1).ItemId
		item2Id := test.AddItemToDatabase(db, sellerId, 2).ItemId

		items, err := queries.GetItems(db)

		if !assert.NoError(t, err) {
			return
		}

		if !assert.Equal(t, 2, len(items)) {
			return
		}
		if !assert.Equal(t, item1Id, items[0].ItemId) {
			return
		}
		if !assert.Equal(t, item2Id, items[1].ItemId) {
			return
		}
	})
}
