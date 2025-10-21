//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSoldItemsQuery(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("No items in existence", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			query := queries.NewGetSoldItemsQuery()
			soldItems, err := query.Execute(db)
			require.NoError(t, err)
			require.Empty(t, soldItems)
		})

		t.Run("Single item", func(t *testing.T) {
			for i := range 20 {
				t.Run(fmt.Sprintf("i = %d", i), func(t *testing.T) {
					setup, db := NewDatabaseFixture(WithDefaultCategories)
					defer setup.Close()

					seller := setup.Seller()
					cashier := setup.Cashier()
					item := setup.Item(seller.UserID, aux.WithDummyData(i), aux.WithHidden(false))
					sale := setup.Sale(cashier.UserID, []models.ID{item.ItemID})

					query := queries.NewGetSoldItemsQuery()
					actualSoldItems, err := query.Execute(db)
					require.NoError(t, err)
					require.Len(t, actualSoldItems, 1)

					actualSoldItem := actualSoldItems[0]
					expectedSoldItem := queries.SoldItem{
						SaleID:          sale.SaleID,
						CashierID:       sale.CashierID,
						TransactionTime: sale.TransactionTime,
						ItemID:          item.ItemID,
						AddedAt:         item.AddedAt,
						Description:     item.Description,
						PriceInCents:    item.PriceInCents,
						ItemCategoryID:  item.CategoryID,
						SellerID:        item.SellerID,
						Donation:        item.Donation,
						Charity:         item.Charity,
						Large:           item.Large,
					}
					require.Equal(t, &expectedSoldItem, actualSoldItem)
				})
			}
		})

		t.Run("One sold item, one unsold item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			soldItem := setup.Item(seller.UserID, aux.WithHidden(false))
			// Make unsold item
			setup.Item(seller.UserID, aux.WithHidden(false))
			sale := setup.Sale(cashier.UserID, []models.ID{soldItem.ItemID})

			query := queries.NewGetSoldItemsQuery()
			actualSoldItems, err := query.Execute(db)
			require.NoError(t, err)
			require.Len(t, actualSoldItems, 1)

			actualSoldItem := actualSoldItems[0]
			expectedSoldItem := queries.SoldItem{
				SaleID:          sale.SaleID,
				CashierID:       sale.CashierID,
				TransactionTime: sale.TransactionTime,
				ItemID:          soldItem.ItemID,
				AddedAt:         soldItem.AddedAt,
				Description:     soldItem.Description,
				PriceInCents:    soldItem.PriceInCents,
				ItemCategoryID:  soldItem.CategoryID,
				SellerID:        soldItem.SellerID,
				Donation:        soldItem.Donation,
				Charity:         soldItem.Charity,
			}
			require.Equal(t, &expectedSoldItem, actualSoldItem)
		})

		t.Run("Two items sold in same sale", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			item1 := setup.Item(seller.UserID, aux.WithHidden(false))
			item2 := setup.Item(seller.UserID, aux.WithHidden(false))
			sale := setup.Sale(cashier.UserID, []models.ID{item1.ItemID, item2.ItemID})

			query := queries.NewGetSoldItemsQuery()
			actualSoldItems, err := query.Execute(db)
			require.NoError(t, err)
			require.Len(t, actualSoldItems, 2)

			expectedSoldItems := []*queries.SoldItem{
				{
					SaleID:          sale.SaleID,
					CashierID:       sale.CashierID,
					TransactionTime: sale.TransactionTime,
					ItemID:          item1.ItemID,
					AddedAt:         item1.AddedAt,
					Description:     item1.Description,
					PriceInCents:    item1.PriceInCents,
					ItemCategoryID:  item1.CategoryID,
					SellerID:        item1.SellerID,
					Donation:        item1.Donation,
					Charity:         item1.Charity,
				},
				{
					SaleID:          sale.SaleID,
					CashierID:       sale.CashierID,
					TransactionTime: sale.TransactionTime,
					ItemID:          item2.ItemID,
					AddedAt:         item2.AddedAt,
					Description:     item2.Description,
					PriceInCents:    item2.PriceInCents,
					ItemCategoryID:  item2.CategoryID,
					SellerID:        item2.SellerID,
					Donation:        item2.Donation,
					Charity:         item2.Charity,
				},
			}
			require.Equal(t, expectedSoldItems, actualSoldItems)
		})

		t.Run("Two items sold in separate sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			item1 := setup.Item(seller.UserID, aux.WithHidden(false))
			item2 := setup.Item(seller.UserID, aux.WithHidden(false))
			sale1 := setup.Sale(cashier.UserID, []models.ID{item1.ItemID})
			sale2 := setup.Sale(cashier.UserID, []models.ID{item2.ItemID})

			query := queries.NewGetSoldItemsQuery()
			actualSoldItems, err := query.Execute(db)
			require.NoError(t, err)
			require.Len(t, actualSoldItems, 2)

			expectedSoldItems := []*queries.SoldItem{
				{
					SaleID:          sale1.SaleID,
					CashierID:       sale1.CashierID,
					TransactionTime: sale1.TransactionTime,
					ItemID:          item1.ItemID,
					AddedAt:         item1.AddedAt,
					Description:     item1.Description,
					PriceInCents:    item1.PriceInCents,
					ItemCategoryID:  item1.CategoryID,
					SellerID:        item1.SellerID,
					Donation:        item1.Donation,
					Charity:         item1.Charity,
				},
				{
					SaleID:          sale2.SaleID,
					CashierID:       sale2.CashierID,
					TransactionTime: sale2.TransactionTime,
					ItemID:          item2.ItemID,
					AddedAt:         item2.AddedAt,
					Description:     item2.Description,
					PriceInCents:    item2.PriceInCents,
					ItemCategoryID:  item2.CategoryID,
					SellerID:        item2.SellerID,
					Donation:        item2.Donation,
					Charity:         item2.Charity,
				},
			}
			require.Equal(t, expectedSoldItems, actualSoldItems)
		})

		t.Run("Same item sold in separate sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			item := setup.Item(seller.UserID, aux.WithHidden(false))
			sale1 := setup.Sale(cashier.UserID, []models.ID{item.ItemID})
			sale2 := setup.Sale(cashier.UserID, []models.ID{item.ItemID})

			query := queries.NewGetSoldItemsQuery()
			actualSoldItems, err := query.Execute(db)
			require.NoError(t, err)
			require.Len(t, actualSoldItems, 2)

			expectedSoldItems := []*queries.SoldItem{
				{
					SaleID:          sale1.SaleID,
					CashierID:       sale1.CashierID,
					TransactionTime: sale1.TransactionTime,
					ItemID:          item.ItemID,
					AddedAt:         item.AddedAt,
					Description:     item.Description,
					PriceInCents:    item.PriceInCents,
					ItemCategoryID:  item.CategoryID,
					SellerID:        item.SellerID,
					Donation:        item.Donation,
					Charity:         item.Charity,
				},
				{
					SaleID:          sale2.SaleID,
					CashierID:       sale2.CashierID,
					TransactionTime: sale2.TransactionTime,
					ItemID:          item.ItemID,
					AddedAt:         item.AddedAt,
					Description:     item.Description,
					PriceInCents:    item.PriceInCents,
					ItemCategoryID:  item.CategoryID,
					SellerID:        item.SellerID,
					Donation:        item.Donation,
					Charity:         item.Charity,
				},
			}
			require.Equal(t, expectedSoldItems, actualSoldItems)
		})
	})
}
