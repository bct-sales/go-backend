//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
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
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			item := setup.Item(seller.UserId, aux.WithHidden(false))
			sale := setup.Sale(cashier.UserId, []models.Id{item.ItemID})

			query := queries.NewGetSoldItemsQuery()
			soldItems, err := query.Execute(db)
			require.NoError(t, err)
			require.Len(t, soldItems, 1)

			actualSoldItem := soldItems[0]
			expectedSoldItem := queries.SoldItem{
				SaleId:          sale.SaleID,
				CashierId:       sale.CashierID,
				TransactionTime: sale.TransactionTime,
				ItemId:          item.ItemID,
				AddedAt:         item.AddedAt,
				Description:     item.Description,
				PriceInCents:    item.PriceInCents,
				ItemCategory:    item.CategoryID,
				SellerId:        item.SellerID,
				Donation:        item.Donation,
				Charity:         item.Charity,
			}
			require.Equal(t, expectedSoldItem, actualSoldItem)
		})

		t.Run("One sold item, one unsold item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			soldItem := setup.Item(seller.UserId, aux.WithHidden(false))
			// Make unsold item
			setup.Item(seller.UserId, aux.WithHidden(false))
			sale := setup.Sale(cashier.UserId, []models.Id{soldItem.ItemID})

			query := queries.NewGetSoldItemsQuery()
			soldItems, err := query.Execute(db)
			require.NoError(t, err)
			require.Len(t, soldItems, 1)

			actualSoldItem := soldItems[0]
			expectedSoldItem := queries.SoldItem{
				SaleId:          sale.SaleID,
				CashierId:       sale.CashierID,
				TransactionTime: sale.TransactionTime,
				ItemId:          soldItem.ItemID,
				AddedAt:         soldItem.AddedAt,
				Description:     soldItem.Description,
				PriceInCents:    soldItem.PriceInCents,
				ItemCategory:    soldItem.CategoryID,
				SellerId:        soldItem.SellerID,
				Donation:        soldItem.Donation,
				Charity:         soldItem.Charity,
			}
			require.Equal(t, expectedSoldItem, actualSoldItem)
		})
	})
}
