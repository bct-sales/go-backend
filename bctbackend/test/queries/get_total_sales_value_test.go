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

func TestGetTotalSalesValue(t *testing.T) {
	t.Run("Zero sales, zero items", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		total, err := queries.GetTotalSalesValue(db)
		require.NoError(t, err)
		require.Equal(t, models.MoneyInCents(0), total)
	})

	t.Run("Zero sales", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		setup.Items(seller.UserId, 10, aux.WithHidden(false))

		total, err := queries.GetTotalSalesValue(db)
		require.NoError(t, err)
		require.Equal(t, models.MoneyInCents(0), total)
	})

	t.Run("Single sale with single item", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()
		items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
		setup.Sale(cashier.UserId, []models.Id{items[0].ItemID})

		total, err := queries.GetTotalSalesValue(db)
		require.NoError(t, err)
		require.Equal(t, items[0].PriceInCents, total)
	})

	t.Run("Single sale with multiple item", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()
		items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
		setup.Sale(cashier.UserId, []models.Id{items[0].ItemID, items[1].ItemID, items[2].ItemID})

		total, err := queries.GetTotalSalesValue(db)
		require.NoError(t, err)
		require.Equal(t, items[0].PriceInCents+items[1].PriceInCents+items[2].PriceInCents, total)
	})

	t.Run("Multiple sales, single cashier, no shared items among sales", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()
		items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
		setup.Sale(cashier.UserId, []models.Id{items[0].ItemID})
		setup.Sale(cashier.UserId, []models.Id{items[1].ItemID})
		setup.Sale(cashier.UserId, []models.Id{items[2].ItemID})

		total, err := queries.GetTotalSalesValue(db)
		require.NoError(t, err)
		require.Equal(t, items[0].PriceInCents+items[1].PriceInCents+items[2].PriceInCents, total)
	})

	t.Run("Multiple sales, single cashier, with shared items among sales", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()
		items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
		setup.Sale(cashier.UserId, []models.Id{items[0].ItemID})
		setup.Sale(cashier.UserId, []models.Id{items[0].ItemID})
		setup.Sale(cashier.UserId, []models.Id{items[0].ItemID})

		total, err := queries.GetTotalSalesValue(db)
		require.NoError(t, err)
		require.Equal(t, items[0].PriceInCents*3, total)
	})

	t.Run("Multiple sales, multiple cashiers, no shared items among sales", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier1 := setup.Cashier()
		cashier2 := setup.Cashier()
		items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
		setup.Sale(cashier1.UserId, []models.Id{items[0].ItemID})
		setup.Sale(cashier2.UserId, []models.Id{items[1].ItemID})
		setup.Sale(cashier2.UserId, []models.Id{items[2].ItemID})
		setup.Sale(cashier2.UserId, []models.Id{items[3].ItemID})

		total, err := queries.GetTotalSalesValue(db)
		require.NoError(t, err)
		require.Equal(t, items[0].PriceInCents+items[1].PriceInCents+items[2].PriceInCents+items[3].PriceInCents, total)
	})

	t.Run("Multiple sales, multiple cashiers, with shared items among sales", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier1 := setup.Cashier()
		cashier2 := setup.Cashier()
		items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
		setup.Sale(cashier1.UserId, []models.Id{items[0].ItemID})
		setup.Sale(cashier1.UserId, []models.Id{items[1].ItemID})
		setup.Sale(cashier2.UserId, []models.Id{items[1].ItemID})
		setup.Sale(cashier2.UserId, []models.Id{items[2].ItemID})
		setup.Sale(cashier2.UserId, []models.Id{items[3].ItemID})

		total, err := queries.GetTotalSalesValue(db)
		require.NoError(t, err)
		require.Equal(t, items[0].PriceInCents+2*items[1].PriceInCents+items[2].PriceInCents+items[3].PriceInCents, total)
	})
}
