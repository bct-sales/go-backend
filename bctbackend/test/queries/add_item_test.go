//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/defs"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestAddItem(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, timestamp := range []models.Timestamp{0, 1000} {
			for _, priceInCents := range []models.MoneyInCents{50, 100} {
				for _, itemCategoryId := range defs.ListCategories() {
					for _, description := range []string{"desc1", "desc2"} {
						for _, sellerId := range []models.Id{1, 2} {
							for _, donation := range []bool{false, true} {
								for _, charity := range []bool{false, true} {
									for _, frozen := range []bool{false, true} {
										for _, hidden := range []bool{false, true} {
											test_name := fmt.Sprintf("timestamp = %d", timestamp)

											t.Run(test_name, func(t *testing.T) {
												setup, db := NewDatabaseFixture()
												defer setup.Close()

												setup.Seller(aux.WithUserId(1))
												setup.Seller(aux.WithUserId(2))

												itemId, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity, frozen, hidden)
												require.NoError(t, err, `Failed to add item: %v`, err)

												{
													itemExists, err := queries.ItemWithIdExists(db, itemId)
													require.NoError(t, err)
													require.True(t, itemExists)
												}

												items := []*models.Item{}
												err = queries.GetItems(db, queries.CollectTo(&items), queries.AllItems)
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
												require.Equal(t, frozen, item.Frozen)
												require.Equal(t, hidden, item.Hidden)
											})
										}
									}
								}
							}
						}
					}
				}
			}
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent seller", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			timestamp := models.NewTimestamp(0)
			description := "description"
			priceInCents := models.NewMoneyInCents(100)
			itemCategoryId := models.NewId(1)
			charity := false
			sellerId := models.NewId(1)
			donation := false
			frozen := false
			hidden := false

			setup.Seller(aux.WithUserId(2))

			_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity, frozen, hidden)
			var noSuchUserError *queries.NoSuchUserError
			require.ErrorAs(t, err, &noSuchUserError)

			count, err := queries.CountItems(db, queries.OnlyVisibleItems)
			require.NoError(t, err)
			require.Equal(t, 0, count)
		})

		t.Run("Nonexistent category", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			timestamp := models.NewTimestamp(0)
			description := "description"
			sellerId := models.NewId(1)
			priceInCents := models.NewMoneyInCents(100)
			charity := false
			donation := false
			frozen := false
			hidden := false
			itemCategoryId := models.NewId(100)

			setup.Seller(aux.WithUserId(1))

			{
				categoryExists, err := queries.CategoryWithIdExists(db, itemCategoryId)
				require.NoError(t, err)
				require.False(t, categoryExists)
			}

			{
				_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity, frozen, hidden)
				var noSuchCategoryError *queries.NoSuchCategoryError
				require.ErrorAs(t, err, &noSuchCategoryError)
			}

			{
				count, err := queries.CountItems(db, queries.OnlyVisibleItems)
				require.NoError(t, err)
				require.Equal(t, 0, count)
			}
		})

		t.Run("Zero price", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			timestamp := models.NewTimestamp(0)
			description := "description"
			itemCategoryId := models.NewId(1)
			charity := false
			seller := setup.Seller()
			donation := false
			frozen := false
			hidden := false
			priceInCents := models.NewMoneyInCents(0)

			{
				_, error := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, seller.UserId, donation, charity, frozen, hidden)

				var invalidPriceError *queries.InvalidPriceError
				require.ErrorAs(t, error, &invalidPriceError)
			}

			{
				count, err := queries.CountItems(db, queries.OnlyVisibleItems)
				require.NoError(t, err)
				require.Equal(t, 0, count)
			}
		})

		t.Run("Negative price", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			timestamp := models.NewTimestamp(0)
			description := "description"
			itemCategoryId := models.NewId(1)
			charity := false
			seller := setup.Seller()
			donation := false
			frozen := false
			hidden := false
			priceInCents := models.NewMoneyInCents(-100)

			{
				_, error := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, seller.UserId, donation, charity, frozen, hidden)
				var invalidPriceError *queries.InvalidPriceError
				require.ErrorAs(t, error, &invalidPriceError)
			}

			{
				count, err := queries.CountItems(db, queries.OnlyVisibleItems)
				require.NoError(t, err)
				require.Equal(t, 0, count)
			}
		})

		t.Run("Cashier owner", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			timestamp := models.NewTimestamp(0)
			description := "description"
			invalidSeller := setup.Cashier()
			priceInCents := models.NewMoneyInCents(100)
			itemCategoryId := models.NewId(1)
			charity := false
			donation := false
			frozen := false
			hidden := false

			{
				_, error := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, invalidSeller.UserId, donation, charity, frozen, hidden)
				var invalidRoleError *queries.InvalidRoleError
				require.ErrorAs(t, error, &invalidRoleError)
			}

			{
				count, err := queries.CountItems(db, queries.OnlyVisibleItems)
				require.NoError(t, err)
				require.Equal(t, 0, count)
			}
		})

		t.Run("Admin owner", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			timestamp := models.NewTimestamp(0)
			description := "description"
			invalidSeller := setup.Admin()
			priceInCents := models.NewMoneyInCents(100)
			itemCategoryId := models.NewId(1)
			charity := false
			donation := false
			frozen := false
			hidden := false

			{
				_, error := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, invalidSeller.UserId, donation, charity, frozen, hidden)
				var invalidRoleError *queries.InvalidRoleError
				require.ErrorAs(t, error, &invalidRoleError)
			}

			{
				count, err := queries.CountItems(db, queries.OnlyVisibleItems)
				require.NoError(t, err)
				require.Equal(t, 0, count)
			}
		})
	})
}
