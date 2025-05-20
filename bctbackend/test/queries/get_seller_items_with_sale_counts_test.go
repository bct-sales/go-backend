//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetSellerItemsWithSaleCounts(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("Zero items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()

			items, err := queries.GetSellerItemsWithSaleCounts(db, seller.UserId)
			require.NoError(t, err)
			require.Equal(t, 0, len(items))
		})

		t.Run("One item in zero sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))

			items, err := queries.GetSellerItemsWithSaleCounts(db, seller.UserId)
			require.NoError(t, err)
			require.Equal(t, 1, len(items))
			require.Equal(t, item.ItemId, items[0].ItemId)
			require.Equal(t, item.Description, items[0].Description)
			require.Equal(t, item.SellerId, items[0].SellerId)
			require.Equal(t, item.AddedAt, items[0].AddedAt)
			require.Equal(t, item.PriceInCents, items[0].PriceInCents)
			require.Equal(t, item.CategoryId, items[0].CategoryId)
			require.Equal(t, item.Charity, items[0].Charity)
			require.Equal(t, item.Donation, items[0].Donation)
			require.Equal(t, item.Frozen, items[0].Frozen)
			require.Equal(t, 0, items[0].SaleCount)
		})

		t.Run("One item in one sale", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.Id{item.ItemId})

			items, err := queries.GetSellerItemsWithSaleCounts(db, seller.UserId)
			require.NoError(t, err)
			require.Equal(t, 1, len(items))
			require.Equal(t, item.ItemId, items[0].ItemId)
			require.Equal(t, 1, items[0].SaleCount)
		})

		t.Run("One item in two sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.Id{item.ItemId})
			setup.Sale(cashier.UserId, []models.Id{item.ItemId})

			items, err := queries.GetSellerItemsWithSaleCounts(db, seller.UserId)
			require.NoError(t, err)
			require.Equal(t, 1, len(items))
			require.Equal(t, item.ItemId, items[0].ItemId)
			require.Equal(t, 2, items[0].SaleCount)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Nonexistent seller", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			invalidSellerId := models.Id(1000)
			setup.RequireNoSuchUser(t, invalidSellerId)

			_, err := queries.GetSellerItemsWithSaleCounts(db, invalidSellerId)
			var errNoSuchUser *queries.NoSuchUserError
			require.ErrorAs(t, err, &errNoSuchUser)
		})
	})
}
