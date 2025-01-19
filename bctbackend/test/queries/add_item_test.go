//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/defs"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
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
									db := OpenInitializedDatabase()
									defer db.Close()

									AddSellerToDatabase(db, WithUserId(1))
									AddSellerToDatabase(db, WithUserId(2))

									itemId, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
									if err != nil {
										t.Fatalf(`Failed to add item: %v`, err)
									}

									itemExists, err := queries.ItemWithIdExists(db, itemId)
									require.NoError(t, err)
									require.True(t, itemExists)

									items, err := queries.GetItems(db)
									require.NoError(t, err)
									require.Equal(t, 1, len(items))

									item := items[0]
									require.Equal(t, timestamp, item.AddedAt)
									require.Equal(t, description, item.Description)
									require.Equal(t, priceInCents, item.PriceInCents)
									require.Equal(t, itemCategoryId, item.CategoryId)
									require.Equal(t, sellerId, sellerId)
									require.Equal(t, donation, item.Donation)
									require.Equal(t, charity, item.Charity)
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
	db := OpenInitializedDatabase()
	defer db.Close()

	timestamp := models.NewTimestamp(0)
	description := "description"
	priceInCents := models.NewMoneyInCents(100)
	itemCategoryId := models.NewId(1)
	charity := false
	sellerId := models.NewId(1)
	donation := false

	AddSellerToDatabase(db, WithUserId(2))

	_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
	var unknownUserError *queries.UnknownUserError
	require.ErrorAs(t, err, &unknownUserError)

	count, err := queries.CountItems(db)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func TestAddItemWithNonexistingCategory(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	timestamp := models.NewTimestamp(0)
	description := "description"
	sellerId := models.NewId(1)
	priceInCents := models.NewMoneyInCents(100)
	charity := false
	donation := false
	itemCategoryId := models.NewId(100)

	AddSellerToDatabase(db, WithUserId(1))

	{
		categoryExists, err := queries.CategoryWithIdExists(db, itemCategoryId)
		require.NoError(t, err)
		require.False(t, categoryExists)
	}

	{
		_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
		var noSuchCategoryError *queries.NoSuchCategoryError
		require.ErrorAs(t, err, &noSuchCategoryError)
	}

	{
		count, err := queries.CountItems(db)
		require.NoError(t, err)
		require.Equal(t, 0, count)
	}
}

func TestAddItemWithZeroPrice(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	timestamp := models.NewTimestamp(0)
	description := "description"
	itemCategoryId := models.NewId(1)
	charity := false
	sellerId := AddSellerToDatabase(db).UserId
	donation := false
	priceInCents := models.NewMoneyInCents(0)

	{
		_, error := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)

		var invalidPriceError *queries.InvalidPriceError
		require.ErrorAs(t, error, &invalidPriceError)
	}

	{
		count, err := queries.CountItems(db)
		require.NoError(t, err)
		require.Equal(t, 0, count)
	}
}

func TestAddItemWithNegativePrice(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	timestamp := models.NewTimestamp(0)
	description := "description"
	itemCategoryId := models.NewId(1)
	charity := false
	sellerId := AddSellerToDatabase(db).UserId
	donation := false
	priceInCents := models.NewMoneyInCents(-100)

	{
		_, error := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
		var invalidPriceError *queries.InvalidPriceError
		require.ErrorAs(t, error, &invalidPriceError)
	}

	{
		count, err := queries.CountItems(db)
		require.NoError(t, err)
		require.Equal(t, 0, count)
	}
}

func TestAddItemWithCashierOwner(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	timestamp := models.NewTimestamp(0)
	description := "description"
	sellerId := AddCashierToDatabase(db).UserId
	priceInCents := models.NewMoneyInCents(100)
	itemCategoryId := models.NewId(1)
	charity := false
	donation := false

	{
		_, error := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
		require.Error(t, error)
	}

	{
		count, err := queries.CountItems(db)
		require.NoError(t, err)
		require.Equal(t, 0, count)
	}
}

func TestAddItemWithAdminOwner(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	timestamp := models.NewTimestamp(0)
	description := "description"
	sellerId := AddAdminToDatabase(db).UserId
	priceInCents := models.NewMoneyInCents(100)
	itemCategoryId := models.NewId(1)
	charity := false
	donation := false

	{
		_, error := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity)
		require.Error(t, error)
	}

	{
		count, err := queries.CountItems(db)
		require.NoError(t, err)
		require.Equal(t, 0, count)
	}
}
