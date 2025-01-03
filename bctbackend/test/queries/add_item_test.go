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

func TestAddItemWithNonexistingSeller(t *testing.T) {
	db := test.OpenInitializedDatabase()

	timestamp := models.NewTimestamp(0)
	description := "description"
	priceInCents := models.NewMoneyInCents(100)
	itemCategoryId := models.NewId(1)
	charity := false
	sellerId := models.NewId(1)
	donation := false

	test.AddSellerWithIdToDatabase(db, 2)

	_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
	var unknownUserError *queries.UnknownUserError
	if !assert.ErrorAs(t, err, &unknownUserError) {
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

func TestAddItemWithNonexistingCategory(t *testing.T) {
	db := test.OpenInitializedDatabase()

	timestamp := models.NewTimestamp(0)
	description := "description"
	sellerId := models.NewId(1)
	priceInCents := models.NewMoneyInCents(100)
	charity := false
	donation := false
	itemCategoryId := models.NewId(100)

	test.AddSellerWithIdToDatabase(db, 1)

	{
		categoryExists, err := queries.CategoryWithIdExists(db, itemCategoryId)

		if !assert.NoError(t, err) {
			return
		}

		if !assert.False(t, categoryExists) {
			return
		}
	}

	{
		_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)

		var noSuchCategoryError *queries.NoSuchCategoryError
		if !assert.ErrorAs(t, err, &noSuchCategoryError) {
			return
		}
	}

	{
		count, err := queries.CountItems(db)

		if !assert.NoError(t, err) {
			return
		}

		if !assert.Equal(t, 0, count) {
			return
		}
	}
}

func TestAddItemWithZeroPrice(t *testing.T) {
	timestamp := models.NewTimestamp(0)
	description := "description"
	itemCategoryId := models.NewId(1)
	charity := false
	db := test.OpenInitializedDatabase()
	sellerId := test.AddSellerToDatabase(db).UserId
	donation := false
	priceInCents := models.NewMoneyInCents(0)

	{
		categoryExists, err := queries.CategoryWithIdExists(db, itemCategoryId)

		if !assert.NoError(t, err) {
			return
		}

		if !assert.True(t, categoryExists) {
			return
		}
	}

	{
		_, error := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)

		if !assert.Error(t, error) {
			return
		}
	}

	{
		count, err := queries.CountItems(db)

		if !assert.NoError(t, err) {
			return
		}

		if !assert.Equal(t, 0, count) {
			return
		}
	}
}

func TestAddItemWithNegativePrice(t *testing.T) {
	timestamp := models.NewTimestamp(0)
	description := "description"
	itemCategoryId := models.NewId(1)
	charity := false
	db := test.OpenInitializedDatabase()
	sellerId := test.AddSellerToDatabase(db).UserId
	donation := false
	priceInCents := models.NewMoneyInCents(-100)

	{
		categoryExists, err := queries.CategoryWithIdExists(db, itemCategoryId)

		if !assert.NoError(t, err) {
			return
		}

		if !assert.True(t, categoryExists) {
			return
		}
	}

	{
		_, error := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)

		if !assert.Error(t, error) {
			return
		}
	}

	{
		count, err := queries.CountItems(db)

		if !assert.NoError(t, err) {
			return
		}

		if !assert.Equal(t, 0, count) {
			return
		}
	}
}

func TestAddItemWithCashierOwner(t *testing.T) {
	db := test.OpenInitializedDatabase()
	timestamp := models.NewTimestamp(0)
	description := "description"
	sellerId := test.AddCashierToDatabase(db).UserId
	priceInCents := models.NewMoneyInCents(100)
	itemCategoryId := models.NewId(1)
	charity := false
	donation := false

	{
		_, error := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)

		if !assert.Error(t, error) {
			return
		}
	}

	{
		count, err := queries.CountItems(db)

		if !assert.NoError(t, err) {
			return
		}

		if !assert.Equal(t, 0, count) {
			return
		}
	}
}

func TestAddItemWithAdminOwner(t *testing.T) {
	db := test.OpenInitializedDatabase()
	timestamp := models.NewTimestamp(0)
	description := "description"
	sellerId := test.AddAdminToDatabase(db).UserId
	priceInCents := models.NewMoneyInCents(100)
	itemCategoryId := models.NewId(1)
	charity := false
	donation := false

	{
		_, error := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)

		if !assert.Error(t, error) {
			return
		}
	}

	{
		count, err := queries.CountItems(db)

		if !assert.NoError(t, err) {
			return
		}

		if !assert.Equal(t, 0, count) {
			return
		}
	}
}
