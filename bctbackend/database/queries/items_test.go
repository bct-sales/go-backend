package queries

import (
	models "bctbackend/database/models"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestAddItem(t *testing.T) {
	for _, timestamp := range []models.Timestamp{0, 1000} {
		for _, priceInCents := range []models.MoneyInCents{50, 100} {
			for _, itemCategoryId := range models.Categories() {
				for _, description := range []string{"desc1", "desc2"} {
					for _, sellerId := range []models.Id{1, 2} {
						for _, donation := range []bool{false, true} {
							for _, charity := range []bool{false, true} {
								test_name := fmt.Sprintf("timestamp = %d", timestamp)

								t.Run(test_name, func(t *testing.T) {
									db := openInitializedDatabase()
									defer db.Close()

									addTestSellerWithId(db, 1)
									addTestSellerWithId(db, 2)

									itemId, err := AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
									if err != nil {
										t.Fatalf(`Failed to add item: %v`, err)
									}

									itemExists, err := ItemWithIdExists(db, itemId)
									if assert.NoError(t, err) {
										assert.True(t, itemExists)
									}

									items, err := GetItems(db)
									assert.NoError(t, err)
									assert.Equal(t, 1, len(items))

									item := items[0]

									assert.Equal(t, timestamp, item.Timestamp)
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

func TestFailingAddItem(t *testing.T) {
	timestamp := models.NewTimestamp(0)
	description := "description"
	priceInCents := models.NewMoneyInCents(100)
	itemCategoryId := models.NewId(1)
	charity := false

	t.Run("Nonexisting owner", func(t *testing.T) {
		db := openInitializedDatabase()
		addTestSellerWithId(db, 2)

		sellerId := models.NewId(1)
		donation := false

		_, error := AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
		assert.Error(t, error)

		count, err := CountItems(db)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Nonexisting category", func(t *testing.T) {
		db := openInitializedDatabase()
		addTestSellerWithId(db, 1)

		sellerId := models.NewId(1)
		donation := false
		itemCategoryId := models.NewId(100)

		assert.False(t, CategoryWithIdExists(db, itemCategoryId))

		_, error := AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
		assert.Error(t, error)

		count, err := CountItems(db)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Invalid price", func(t *testing.T) {
		db := openInitializedDatabase()
		sellerId := addTestSeller(db)
		donation := false
		itemCategoryId := models.Shoes
		priceInCents := models.NewMoneyInCents(0)

		assert.True(t, CategoryWithIdExists(db, itemCategoryId))

		_, error := AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
		assert.Error(t, error)

		count, err := CountItems(db)
		if assert.NoError(t, err) {
			assert.Equal(t, 0, count)
		}
	})

	t.Run("Cashier owner", func(t *testing.T) {
		db := openInitializedDatabase()
		sellerId := addTestCashier(db)
		donation := false

		_, error := AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
		assert.Error(t, error)
	})

	t.Run("Admin owner", func(t *testing.T) {
		db := openInitializedDatabase()
		sellerId := addTestAdmin(db)
		donation := false

		_, error := AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
		assert.Error(t, error)
	})
}

func TestGetItems(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		db := openInitializedDatabase()
		defer db.Close()

		items, err := GetItems(db)
		if assert.NoError(t, err) {
			assert.Equal(t, 0, len(items))
		}
	})

	t.Run("One item", func(t *testing.T) {
		db := openInitializedDatabase()
		defer db.Close()

		sellerId := addTestSeller(db)
		itemId := addTestItem(db, sellerId, 1)

		items, err := GetItems(db)
		if assert.NoError(t, err) {
			assert.Equal(t, 1, len(items))
			assert.Equal(t, itemId, items[0].ItemId)
		}
	})

	t.Run("Two items", func(t *testing.T) {
		db := openInitializedDatabase()
		defer db.Close()

		sellerId := addTestSeller(db)
		item1Id := addTestItem(db, sellerId, 1)
		item2Id := addTestItem(db, sellerId, 2)

		items, err := GetItems(db)
		if assert.NoError(t, err) {
			assert.Equal(t, 2, len(items))
			assert.Equal(t, item1Id, items[0].ItemId)
			assert.Equal(t, item2Id, items[1].ItemId)
		}
	})
}

func TestGetItemWithId(t *testing.T) {
	t.Run("Nonexisting item", func(t *testing.T) {
		db := openInitializedDatabase()
		defer db.Close()

		itemId := models.NewId(1)
		_, err := GetItemWithId(db, itemId)

		var itemNotFoundError *ItemNotFoundError
		assert.ErrorAs(t, err, &itemNotFoundError)
	})

	t.Run("Existing item", func(t *testing.T) {
		db := openInitializedDatabase()
		defer db.Close()

		sellerId := addTestSeller(db)
		itemId := addTestItem(db, sellerId, 1)

		item, err := GetItemWithId(db, itemId)
		if assert.NoError(t, err) {
			assert.Equal(t, itemId, item.ItemId)
		}
	})
}

func TestRemoveItemWithId(t *testing.T) {
	t.Run("Nonexisting item", func(t *testing.T) {
		db := openInitializedDatabase()
		defer db.Close()

		itemId := models.NewId(1)
		err := RemoveItemWithId(db, itemId)

		var itemNotFoundError *ItemNotFoundError
		assert.ErrorAs(t, err, &itemNotFoundError)
	})

	t.Run("Existing item", func(t *testing.T) {
		db := openInitializedDatabase()
		defer db.Close()

		sellerId := addTestSeller(db)
		itemId := addTestItem(db, sellerId, 1)

		err := RemoveItemWithId(db, itemId)
		if assert.NoError(t, err) {
			itemExists, err := ItemWithIdExists(db, itemId)

			if assert.NoError(t, err) {
				assert.False(t, itemExists)
			}
		}
	})

	t.Run("Existing item with sale", func(t *testing.T) {
		db := openInitializedDatabase()
		defer db.Close()

		sellerId := addTestSeller(db)
		cashierId := addTestCashier(db)
		itemId := addTestItem(db, sellerId, 1)

		addTestSale(db, cashierId, []models.Id{itemId})

		err := RemoveItemWithId(db, itemId)
		assert.Error(t, err)

		itemExists, err := ItemWithIdExists(db, itemId)
		if assert.NoError(t, err) {
			assert.True(t, itemExists)
		}
	})
}

func TestCountItems(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		db := openInitializedDatabase()
		defer db.Close()

		count, err := CountItems(db)
		if assert.NoError(t, err) {
			assert.Equal(t, 0, count)
		}
	})

	t.Run("One item", func(t *testing.T) {
		db := openInitializedDatabase()
		defer db.Close()

		sellerId := addTestSeller(db)
		addTestItem(db, sellerId, 1)

		count, err := CountItems(db)
		if assert.NoError(t, err) {
			assert.Equal(t, 1, count)
		}
	})

	t.Run("Two items", func(t *testing.T) {
		db := openInitializedDatabase()
		defer db.Close()

		sellerId := addTestSeller(db)
		addTestItem(db, sellerId, 1)
		addTestItem(db, sellerId, 2)

		count, err := CountItems(db)
		if assert.NoError(t, err) {
			assert.Equal(t, 2, count)
		}
	})
}
