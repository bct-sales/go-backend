//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/defs"
	"bctbackend/test"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestAddItemToDatabase(t *testing.T) {
	for _, timestamp := range []models.Timestamp{0, 1000} {
		for _, priceInCents := range []models.MoneyInCents{50, 100} {
			for _, itemCategoryId := range defs.ListCategories() {
				for _, description := range []string{"desc1", "desc2"} {
					for _, sellerId := range []models.Id{1, 2} {
						for _, donation := range []bool{false, true} {
							for _, charity := range []bool{false, true} {
								test_name := fmt.Sprintf("timestamp = %d", timestamp)

								t.Run(test_name, func(t *testing.T) {
									db := test.OpenInitializedDatabase()
									defer db.Close()

									test.AddSellerWithIdToDatabase(db, 1)
									test.AddSellerWithIdToDatabase(db, 2)

									itemId, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
									if err != nil {
										t.Fatalf(`Failed to add item: %v`, err)
									}

									itemExists, err := queries.ItemWithIdExists(db, itemId)
									if assert.NoError(t, err) {
										assert.True(t, itemExists)
									}

									items, err := queries.GetItems(db)
									assert.NoError(t, err)
									assert.Equal(t, 1, len(items))

									item := items[0]

									assert.Equal(t, timestamp, item.AddedAt)
									assert.Equal(t, description, item.Description)
									assert.Equal(t, priceInCents, item.PriceInCents)
									assert.Equal(t, itemCategoryId, item.CategoryId)
									assert.Equal(t, sellerId, sellerId)
									assert.Equal(t, donation, item.Donation)
									assert.Equal(t, charity, item.Charity)
								})
							}
						}
					}
				}
			}
		}
	}
}

func TestFailingAddItemToDatabase(t *testing.T) {
	timestamp := models.NewTimestamp(0)
	description := "description"
	priceInCents := models.NewMoneyInCents(100)
	itemCategoryId := models.NewId(1)
	charity := false

	t.Run("Nonexisting owner", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		test.AddSellerWithIdToDatabase(db, 2)

		sellerId := models.NewId(1)
		donation := false

		_, error := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
		assert.Error(t, error)

		count, err := queries.CountItems(db)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Nonexisting category", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		test.AddSellerWithIdToDatabase(db, 1)

		sellerId := models.NewId(1)
		donation := false
		itemCategoryId := models.NewId(100)

		assert.False(t, queries.CategoryWithIdExists(db, itemCategoryId))

		_, error := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
		assert.Error(t, error)

		count, err := queries.CountItems(db)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Invalid price", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		sellerId := test.AddSellerToDatabase(db).UserId
		donation := false
		itemCategoryId := defs.Shoes
		priceInCents := models.NewMoneyInCents(0)

		assert.True(t, queries.CategoryWithIdExists(db, itemCategoryId))

		_, error := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
		assert.Error(t, error)

		count, err := queries.CountItems(db)
		if assert.NoError(t, err) {
			assert.Equal(t, 0, count)
		}
	})

	t.Run("Cashier owner", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		sellerId := test.AddCashierToDatabase(db).UserId
		donation := false

		_, error := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
		assert.Error(t, error)
	})

	t.Run("Admin owner", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		sellerId := test.AddAdminToDatabase(db).UserId
		donation := false

		_, error := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
		assert.Error(t, error)
	})
}

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
