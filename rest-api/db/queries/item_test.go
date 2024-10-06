package queries

import (
	models "bctrest/db/models"
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
					for _, ownerId := range []models.Id{1, 2} {
						for _, recipientId := range []models.Id{1, 2} {
							for _, charity := range []bool{false, true} {
								test_name := fmt.Sprintf("timestamp = %d", timestamp)

								t.Run(test_name, func(t *testing.T) {
									db := openInitializedDatabase()
									addSeller(db, 1)
									addSeller(db, 2)

									if err := AddItem(db, timestamp, description, priceInCents, itemCategoryId, ownerId, recipientId, charity); err != nil {
										t.Fatalf(`Failed to add item: %v`, err)
									}

									items, err := GetItems(db)
									assert.NoError(t, err)
									assert.Equal(t, 1, len(items))

									item := items[0]

									assert.Equal(t, timestamp, item.Timestamp)
									assert.Equal(t, description, item.Description)
									assert.Equal(t, priceInCents, item.PriceInCents)
									assert.Equal(t, itemCategoryId, item.CategoryId)
									assert.Equal(t, ownerId, ownerId)
									assert.Equal(t, recipientId, item.RecipientId)
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
		addSeller(db, 2)

		var ownerId models.Id = 1
		var recipientId models.Id = 2

		assert.Error(t, AddItem(db, timestamp, description, priceInCents, itemCategoryId, ownerId, recipientId, charity))

		count, err := CountItems(db)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Nonexisting recipient", func(t *testing.T) {
		db := openInitializedDatabase()
		addSeller(db, 1)

		var ownerId models.Id = 1
		var recipientId models.Id = 2

		assert.Error(t, AddItem(db, timestamp, description, priceInCents, itemCategoryId, ownerId, recipientId, charity))

		count, err := CountItems(db)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Nonexisting category", func(t *testing.T) {
		db := openInitializedDatabase()
		addSeller(db, 1)

		var ownerId models.Id = 1
		var recipientId models.Id = 1
		var itemCategoryId models.Id = 100

		categoryExists := CategoryWithIdExists(db, itemCategoryId)
		assert.False(t, categoryExists)
		assert.Error(t, AddItem(db, timestamp, description, priceInCents, itemCategoryId, ownerId, recipientId, charity))

		count, err := CountItems(db)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}
