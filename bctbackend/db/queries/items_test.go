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

									addTestSeller(db, 1)
									addTestSeller(db, 2)

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
	var timestamp models.Timestamp = 0
	var description string = "description"
	var priceInCents models.MoneyInCents = 100
	var itemCategoryId models.Id = 1
	var charity bool = false

	t.Run("Nonexisting owner", func(t *testing.T) {
		db := openInitializedDatabase()
		addTestSeller(db, 2)

		var sellerId models.Id = 1
		donation := false

		_, error := AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
		assert.Error(t, error)

		count, err := CountItems(db)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Nonexisting category", func(t *testing.T) {
		db := openInitializedDatabase()
		addTestSeller(db, 1)

		var ownerId models.Id = 1
		donation := false
		var itemCategoryId models.Id = 100

		categoryExists := CategoryWithIdExists(db, itemCategoryId)
		assert.False(t, categoryExists)

		_, error := AddItem(db, timestamp, description, priceInCents, itemCategoryId, ownerId, donation, charity)
		assert.Error(t, error)

		count, err := CountItems(db)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}
