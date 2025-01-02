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

func TestGetItems(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		items, err := queries.GetItems(db)
		if assert.NoError(t, err) {
			assert.Equal(t, 0, len(items))
		}
	})

	t.Run("One item", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId
		itemId := test.AddItemToDatabase(db, sellerId, 1).ItemId

		items, err := queries.GetItems(db)
		if assert.NoError(t, err) {
			assert.Equal(t, 1, len(items))
			assert.Equal(t, itemId, items[0].ItemId)
		}
	})

	t.Run("Two items", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId
		item1Id := test.AddItemToDatabase(db, sellerId, 1).ItemId
		item2Id := test.AddItemToDatabase(db, sellerId, 2).ItemId

		items, err := queries.GetItems(db)
		if assert.NoError(t, err) {
			assert.Equal(t, 2, len(items))
			assert.Equal(t, item1Id, items[0].ItemId)
			assert.Equal(t, item2Id, items[1].ItemId)
		}
	})
}

func TestGetItemWithId(t *testing.T) {
	t.Run("Nonexisting item", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		itemId := models.NewId(1)
		_, err := queries.GetItemWithId(db, itemId)

		var itemNotFoundError *queries.ItemNotFoundError
		assert.ErrorAs(t, err, &itemNotFoundError)
	})

	t.Run("Existing item", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId
		itemId := test.AddItemToDatabase(db, sellerId, 1).ItemId

		item, err := queries.GetItemWithId(db, itemId)
		if assert.NoError(t, err) {
			assert.Equal(t, itemId, item.ItemId)
		}
	})
}

func TestRemoveItemWithId(t *testing.T) {
	t.Run("Nonexisting item", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		itemId := models.NewId(1)
		err := queries.RemoveItemWithId(db, itemId)

		var itemNotFoundError *queries.ItemNotFoundError
		assert.ErrorAs(t, err, &itemNotFoundError)
	})

	t.Run("Existing item", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId
		itemId := test.AddItemToDatabase(db, sellerId, 1).ItemId

		err := queries.RemoveItemWithId(db, itemId)
		if assert.NoError(t, err) {
			itemExists, err := queries.ItemWithIdExists(db, itemId)

			if assert.NoError(t, err) {
				assert.False(t, itemExists)
			}
		}
	})

	t.Run("Existing item with sale", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId
		cashierId := test.AddCashierToDatabase(db).UserId
		itemId := test.AddItemToDatabase(db, sellerId, 1).ItemId

		test.AddSaleToDatabase(db, cashierId, []models.Id{itemId})

		err := queries.RemoveItemWithId(db, itemId)
		assert.Error(t, err)

		itemExists, err := queries.ItemWithIdExists(db, itemId)
		if assert.NoError(t, err) {
			assert.True(t, itemExists)
		}
	})
}

func TestCountItems(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		count, err := queries.CountItems(db)
		if assert.NoError(t, err) {
			assert.Equal(t, 0, count)
		}
	})

	t.Run("One item", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId
		test.AddItemToDatabase(db, sellerId, 1)

		count, err := queries.CountItems(db)
		if assert.NoError(t, err) {
			assert.Equal(t, 1, count)
		}
	})

	t.Run("Two items", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

		sellerId := test.AddSellerToDatabase(db).UserId
		test.AddItemToDatabase(db, sellerId, 1)
		test.AddItemToDatabase(db, sellerId, 2)

		count, err := queries.CountItems(db)
		if assert.NoError(t, err) {
			assert.Equal(t, 2, count)
		}
	})
}
