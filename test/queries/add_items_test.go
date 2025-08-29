//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"

	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Add zero items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			callback := func(addItem queries.AddItemFunction) {}

			err := queries.AddItems(db, callback)
			require.NoError(t, err)

			itemStatistics, err := queries.GetItemStatistics(db, queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, 0, itemStatistics.ItemCount)
			require.Equal(t, models.MoneyInCents(0), itemStatistics.TotalValueInCents)
		})

		t.Run("Add single item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()

			addedAt := models.Timestamp(100)
			description := "Black T-Shirt"
			priceInCents := models.MoneyInCents(1500)
			categoryID := helpers.CategoryId_Clothing56_62
			sellerID := seller.UserId
			donation := false
			charity := false
			frozen := false
			hidden := false
			callback := func(addItem queries.AddItemFunction) {
				addItem(addedAt, description, priceInCents, categoryID, sellerID, donation, charity, frozen, hidden)
			}

			err := queries.AddItems(db, callback)
			require.NoError(t, err)

			itemStatistics, err := queries.GetItemStatistics(db, queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, 1, itemStatistics.ItemCount)
			require.Equal(t, priceInCents, itemStatistics.TotalValueInCents)

			itemIds, err := queries.GetItemIds(db)
			require.NoError(t, err)
			require.Len(t, itemIds, 1)

			itemId := itemIds[0]
			item, err := queries.GetItemWithId(db, itemId)
			require.NoError(t, err)
			require.Equal(t, itemId, item.ItemID)
			require.Equal(t, addedAt, item.AddedAt)
			require.Equal(t, description, item.Description)
			require.Equal(t, priceInCents, item.PriceInCents)
			require.Equal(t, categoryID, item.CategoryID)
			require.Equal(t, sellerID, item.SellerID)
			require.Equal(t, donation, item.Donation)
			require.Equal(t, charity, item.Charity)
			require.Equal(t, frozen, item.Frozen)
			require.Equal(t, hidden, item.Hidden)
		})

		t.Run("Add multiple items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()

			itemCount := 15
			totalPrice := models.MoneyInCents(0)
			callback := func(addItem queries.AddItemFunction) {
				for i := range itemCount {
					addedAt := models.Timestamp(100)
					description := fmt.Sprintf("Item %d", i)
					priceInCents := models.MoneyInCents(50 * (i + 1))
					categoryID := helpers.CategoryId_Clothing56_62
					sellerID := seller.UserId
					donation := i%2 == 0
					charity := i%3 == 0
					frozen := false
					hidden := false

					addItem(addedAt, description, priceInCents, categoryID, sellerID, donation, charity, frozen, hidden)
					totalPrice += priceInCents
				}
			}

			err := queries.AddItems(db, callback)
			require.NoError(t, err)

			itemStatistics, err := queries.GetItemStatistics(db, queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, itemCount, itemStatistics.ItemCount)
			require.Equal(t, totalPrice, itemStatistics.TotalValueInCents)
		})
	})
}
