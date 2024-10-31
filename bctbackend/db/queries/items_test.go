package queries

import (
	models "bctbackend/db/models"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestAddItem(t *testing.T) {
	for _, timestamp := range []models.Timestamp{0, 1000} {
		for _, priceInCents := range []models.MoneyInCents{50, 100} {
			for _, itemCategoryId := range []models.Id{1, 2} {
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

									assert.True(t, ItemWithIdExists(db, itemId))

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
}
