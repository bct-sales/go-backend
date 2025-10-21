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
				for _, itemCategoryID := range defaultCategoryKeys {
					for _, description := range []string{"desc1", "desc2"} {
						for _, sellerID := range []models.ID{1, 2} {
							for _, donation := range []bool{false, true} {
								for _, charity := range []bool{false, true} {
									for _, frozen := range []bool{false, true} {
										for _, hidden := range []bool{false, true} {
											for _, large := range []bool{false, true} {
												test_name := fmt.Sprintf("timestamp = %d, priceInCents = %d, category = %d, seller = %d, donation = %v, frozen = %v, hidden = %v, large = %v", timestamp, priceInCents, itemCategoryID, sellerID, donation, frozen, hidden, large)

												if !hidden || !frozen {
													t.Run(test_name, func(t *testing.T) {
														setup, db := NewDatabaseFixture(WithDefaultCategories)
														defer setup.Close()

														setup.Seller(aux.WithUserID(1))
														setup.Seller(aux.WithUserID(2))

														itemID, addItemErr := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryID, sellerID, donation, charity, frozen, hidden, large)
														require.NoError(t, addItemErr, `Failed to add item: %v`, addItemErr)

														{
															itemExists, err := queries.ItemWithIDExists(db, itemID)
															require.NoError(t, err)
															require.True(t, itemExists)
														}

														items := []*models.Item{}
														query := queries.NewGetItemsQuery()
														queryErr := query.Execute(db, queries.CollectTo(&items))
														require.NoError(t, queryErr)
														require.Equal(t, 1, len(items))

														actualItem := items[0]
														expectedItem := models.Item{
															ItemID:       actualItem.ItemID,
															AddedAt:      timestamp,
															Description:  description,
															PriceInCents: priceInCents,
															CategoryID:   itemCategoryID,
															SellerID:     sellerID,
															Donation:     donation,
															Charity:      charity,
															Frozen:       frozen,
															Hidden:       hidden,
															Large:        large,
														}
														require.Equal(t, &expectedItem, actualItem)
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
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent seller", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			timestamp := models.Timestamp(0)
			description := "description"
			priceInCents := models.MoneyInCents(100)
			itemCategoryID := helpers.CategoryID_Clothing140_152
			charity := false
			sellerID := models.ID(1)
			donation := false
			frozen := false
			hidden := false
			large := false

			setup.Seller(aux.WithUserID(2))

			_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryID, sellerID, donation, charity, frozen, hidden, large)
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)

			query := queries.NewGetItemStatisticsQuery()
			query.WithHidden(false)
			itemStatistics, err := query.Execute(db)
			require.NoError(t, err)
			require.Equal(t, 0, itemStatistics.ItemCount)
			require.Equal(t, models.MoneyInCents(0), itemStatistics.TotalValueInCents)
		})

		t.Run("Nonexistent category", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			timestamp := models.Timestamp(0)
			description := "description"
			sellerID := models.ID(1)
			priceInCents := models.MoneyInCents(100)
			charity := false
			donation := false
			frozen := false
			hidden := false
			large := false
			itemCategoryID := models.ID(100)

			setup.Seller(aux.WithUserID(1))

			{
				categoryExists, err := queries.CategoryWithIDExists(db, itemCategoryID)
				require.NoError(t, err)
				require.False(t, categoryExists)
			}

			{
				_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryID, sellerID, donation, charity, frozen, hidden, large)
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchCategory)
			}

			{
				query := queries.NewGetItemStatisticsQuery()
				query.WithHidden(false)
				itemStatistics, err := query.Execute(db)
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
			itemCategoryID := helpers.CategoryID_Toys
			charity := false
			seller := setup.Seller()
			donation := false
			frozen := false
			hidden := false
			large := false
			priceInCents := models.MoneyInCents(0)

			{
				_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryID, seller.UserID, donation, charity, frozen, hidden, large)
				requireDatabaseWrappedError(t, err, dberr.ErrInvalidPrice)
			}

			{
				query := queries.NewGetItemStatisticsQuery()
				query.WithHidden(false)
				itemStatistics, err := query.Execute(db)
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
			itemCategoryID := helpers.CategoryID_Toys
			charity := false
			seller := setup.Seller()
			donation := false
			frozen := false
			hidden := false
			large := false
			priceInCents := models.MoneyInCents(-100)

			_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryID, seller.UserID, donation, charity, frozen, hidden, large)
			requireDatabaseWrappedError(t, err, dberr.ErrInvalidPrice)

			query := queries.NewGetItemStatisticsQuery()
			query.WithHidden(false)
			itemStatistics, err := query.Execute(db)
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
			itemCategoryID := helpers.CategoryID_Toys
			charity := false
			donation := false
			frozen := false
			hidden := false
			large := false

			_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryID, invalidSeller.UserID, donation, charity, frozen, hidden, large)
			requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)

			{
				query := queries.NewGetItemStatisticsQuery()
				query.WithHidden(false)
				itemStatistics, err := query.Execute(db)
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
			itemCategoryID := helpers.CategoryID_Toys
			charity := false
			donation := false
			frozen := false
			hidden := false
			large := false

			{
				_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryID, invalidSeller.UserID, donation, charity, frozen, hidden, large)
				requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
			}

			{
				query := queries.NewGetItemStatisticsQuery()
				query.WithHidden(false)
				itemStatistics, err := query.Execute(db)
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
			itemCategoryID := helpers.CategoryID_Toys
			charity := false
			donation := false
			frozen := true
			hidden := true
			large := false

			{
				_, err := queries.AddItem(db, timestamp, description, priceInCents, itemCategoryID, seller.UserID, donation, charity, frozen, hidden, large)
				requireDatabaseWrappedError(t, err, dberr.ErrHiddenFrozenItem)
			}

			{
				query := queries.NewGetItemStatisticsQuery()
				query.WithHidden(false)
				itemStatistics, err := query.Execute(db)
				require.NoError(t, err)
				require.Equal(t, 0, itemStatistics.ItemCount)
				require.Equal(t, models.MoneyInCents(0), itemStatistics.TotalValueInCents)
			}
		})
	})
}
