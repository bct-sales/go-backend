//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/test/helpers"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"maps"
	"slices"

	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddItem(t *testing.T) {
	defaultCategoryNameTable := aux.DefaultCategoryNameTable()
	defaultCategoryKeys := slices.Collect(maps.Keys(defaultCategoryNameTable))

	t.Run("Success", func(t *testing.T) {
		for _, timestamp := range []models.Timestamp{0, 1000} {
			for _, priceInCents := range []models.MoneyInCents{50, 100} {
				for _, itemCategoryId := range defaultCategoryKeys {
					for _, description := range []string{"desc1", "desc2"} {
						for _, sellerId := range []models.ID{1, 2} {
							for _, donation := range []bool{false, true} {
								for _, charity := range []bool{false, true} {
									for _, frozen := range []bool{false, true} {
										for _, hidden := range []bool{false, true} {
											test_name := fmt.Sprintf("timestamp = %d", timestamp)

											if !hidden || !frozen {
												t.Run(test_name, func(t *testing.T) {
													setup, db := NewDatabaseFixture(WithDefaultCategories)
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
													err = queries.GetItems(db, queries.CollectTo(&items), queries.AllItems, queries.AllRows())
													require.NoError(t, err)
													require.Equal(t, 1, len(items))

													item := items[0]
													require.Equal(t, timestamp, item.AddedAt)
													require.Equal(t, description, item.Description)
													require.Equal(t, priceInCents, item.PriceInCents)
													require.Equal(t, itemCategoryId, item.CategoryID)
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
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent seller", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			timestamp := models.Timestamp(0)
			description := "description"
			priceInCents := models.MoneyInCents(100)
			itemCategoryId := helpers.CategoryId_Clothing140_152
			charity := false
			sellerId := models.ID(1)
			donation := false
			frozen := false
			hidden := false

			setup.Seller(aux.WithUserId(2))

			_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity, frozen, hidden)
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)

			itemStatistics, err := queries.GetItemStatistics(db, queries.OnlyVisibleItems)
			require.NoError(t, err)
			require.Equal(t, 0, itemStatistics.ItemCount)
			require.Equal(t, models.MoneyInCents(0), itemStatistics.TotalValueInCents)
		})

		t.Run("Nonexistent category", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			timestamp := models.Timestamp(0)
			description := "description"
			sellerId := models.ID(1)
			priceInCents := models.MoneyInCents(100)
			charity := false
			donation := false
			frozen := false
			hidden := false
			itemCategoryId := models.ID(100)

			setup.Seller(aux.WithUserId(1))

			{
				categoryExists, err := queries.CategoryWithIDExists(db, itemCategoryId)
				require.NoError(t, err)
				require.False(t, categoryExists)
			}

			{
				_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, sellerId, donation, charity, frozen, hidden)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchCategory)
			}

			{
				itemStatistics, err := queries.GetItemStatistics(db, queries.OnlyVisibleItems)
				require.NoError(t, err)
				require.Equal(t, 0, itemStatistics.ItemCount)
				require.Equal(t, models.MoneyInCents(0), itemStatistics.TotalValueInCents)
			}
		})

		t.Run("Zero price", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			timestamp := models.Timestamp(0)
			description := "description"
			itemCategoryId := helpers.CategoryId_Toys
			charity := false
			seller := setup.Seller()
			donation := false
			frozen := false
			hidden := false
			priceInCents := models.MoneyInCents(0)

			{
				_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, seller.UserID, donation, charity, frozen, hidden)
				requireDatabaseWrappedError(t, err, dberr.ErrInvalidPrice)
			}

			{
				itemStatistics, err := queries.GetItemStatistics(db, queries.OnlyVisibleItems)
				require.NoError(t, err)
				require.Equal(t, 0, itemStatistics.ItemCount)
				require.Equal(t, models.MoneyInCents(0), itemStatistics.TotalValueInCents)
			}
		})

		t.Run("Negative price", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			timestamp := models.Timestamp(0)
			description := "description"
			itemCategoryId := helpers.CategoryId_Toys
			charity := false
			seller := setup.Seller()
			donation := false
			frozen := false
			hidden := false
			priceInCents := models.MoneyInCents(-100)

			_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, seller.UserID, donation, charity, frozen, hidden)
			requireDatabaseWrappedError(t, err, dberr.ErrInvalidPrice)

			itemStatistics, err := queries.GetItemStatistics(db, queries.OnlyVisibleItems)
			require.NoError(t, err)
			require.Equal(t, 0, itemStatistics.ItemCount)
			require.Equal(t, models.MoneyInCents(0), itemStatistics.TotalValueInCents)
		})

		t.Run("Cashier owner", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			timestamp := models.Timestamp(0)
			description := "description"
			invalidSeller := setup.Cashier()
			priceInCents := models.MoneyInCents(100)
			itemCategoryId := helpers.CategoryId_Toys
			charity := false
			donation := false
			frozen := false
			hidden := false

			_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, invalidSeller.UserID, donation, charity, frozen, hidden)
			requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)

			{
				itemStatistics, err := queries.GetItemStatistics(db, queries.OnlyVisibleItems)
				require.NoError(t, err)
				require.Equal(t, 0, itemStatistics.ItemCount)
				require.Equal(t, models.MoneyInCents(0), itemStatistics.TotalValueInCents)
			}
		})

		t.Run("Admin owner", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			timestamp := models.Timestamp(0)
			description := "description"
			invalidSeller := setup.Admin()
			priceInCents := models.MoneyInCents(100)
			itemCategoryId := helpers.CategoryId_Toys
			charity := false
			donation := false
			frozen := false
			hidden := false

			{
				_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, invalidSeller.UserID, donation, charity, frozen, hidden)
				requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
			}

			{
				itemStatistics, err := queries.GetItemStatistics(db, queries.OnlyVisibleItems)
				require.NoError(t, err)
				require.Equal(t, 0, itemStatistics.ItemCount)
				require.Equal(t, models.MoneyInCents(0), itemStatistics.TotalValueInCents)
			}
		})

		t.Run("Hidden frozen item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			timestamp := models.Timestamp(0)
			description := "description"
			seller := setup.Seller()
			priceInCents := models.MoneyInCents(100)
			itemCategoryId := helpers.CategoryId_Toys
			charity := false
			donation := false
			frozen := true
			hidden := true

			{
				_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryId, seller.UserID, donation, charity, frozen, hidden)
				requireDatabaseWrappedError(t, err, dberr.ErrHiddenFrozenItem)
			}

			{
				itemStatistics, err := queries.GetItemStatistics(db, queries.OnlyVisibleItems)
				require.NoError(t, err)
				require.Equal(t, 0, itemStatistics.ItemCount)
				require.Equal(t, models.MoneyInCents(0), itemStatistics.TotalValueInCents)
			}
		})
	})
}
