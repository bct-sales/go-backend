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

func TestAddItemWithNonexistingUser(t *testing.T) {
	db := test.OpenInitializedDatabase()

	timestamp := models.NewTimestamp(0)
	description := "description"
	priceInCents := models.NewMoneyInCents(100)
	itemCategoryId := models.NewId(1)
	charity := false
	sellerId := models.NewId(1)
	donation := false

	test.AddSellerWithIdToDatabase(db, 2)

	_, error := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
	if !assert.Error(t, error) {
		return
	}

	count, err := queries.CountItems(db)
	if !assert.NoError(t, err) {
		return
	}

	if !assert.Equal(t, 0, count) {
		return
	}
}

func TestFailingAddItemToDatabase(t *testing.T) {
	timestamp := models.NewTimestamp(0)
	description := "description"
	priceInCents := models.NewMoneyInCents(100)
	itemCategoryId := models.NewId(1)
	charity := false

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
