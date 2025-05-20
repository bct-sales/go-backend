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

func createTotalMap(categories []*models.ItemCategory) map[models.Id]models.MoneyInCents {
	totalMap := make(map[models.Id]models.MoneyInCents)
	for _, category := range categories {
		totalMap[category.CategoryId] = models.MoneyInCents(0)
	}
	return totalMap
}

func TestGetSalesOverview(t *testing.T) {
	t.Run("Zero items", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		categories, err := queries.GetCategories(db)
		require.NoError(t, err)

		categorySaleTotals, err := queries.GetSalesOverview(db)
		t.Log(categorySaleTotals)
		require.NoError(t, err)
		require.Equal(t, len(categories), len(categorySaleTotals))

		for categoryIndex, category := range categories {
			require.Equal(t, category.CategoryId, categorySaleTotals[categoryIndex].CategoryId)
			require.Equal(t, category.Name, categorySaleTotals[categoryIndex].CategoryName)
			require.Equal(t, models.MoneyInCents(0), categorySaleTotals[categoryIndex].TotalInCents)
		}
	})

	t.Run("One item, zero sales", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		categories, err := queries.GetCategories(db)
		require.NoError(t, err)

		seller := setup.Seller()
		setup.Item(seller.UserId, aux.WithItemCategory(categories[0].CategoryId), aux.WithDummyData(1), aux.WithHidden(false))

		categorySaleTotals, err := queries.GetSalesOverview(db)
		t.Log(categorySaleTotals)
		require.NoError(t, err)
		require.Equal(t, len(categories), len(categorySaleTotals))

		for categoryIndex, category := range categories {
			require.Equal(t, category.CategoryId, categorySaleTotals[categoryIndex].CategoryId)
			require.Equal(t, category.Name, categorySaleTotals[categoryIndex].CategoryName)
			require.Equal(t, models.MoneyInCents(0), categorySaleTotals[categoryIndex].TotalInCents)
		}
	})

	t.Run("One sold item", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		categories, err := queries.GetCategories(db)
		require.NoError(t, err)

		totals := createTotalMap(categories)

		seller := setup.Seller()
		item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithItemCategory(categories[0].CategoryId), aux.WithHidden(false))
		totals[item.CategoryId] += item.PriceInCents

		cashier := setup.Cashier()
		setup.Sale(cashier.UserId, []models.Id{item.ItemId})

		categorySaleTotals, err := queries.GetSalesOverview(db)
		t.Log(categorySaleTotals)
		require.NoError(t, err)
		require.Equal(t, len(categories), len(categorySaleTotals))

		for categoryIndex, category := range categories {
			require.Equal(t, category.CategoryId, categorySaleTotals[categoryIndex].CategoryId)
			require.Equal(t, category.Name, categorySaleTotals[categoryIndex].CategoryName)
			require.Equal(t, totals[category.CategoryId], categorySaleTotals[categoryIndex].TotalInCents)
		}
	})

	t.Run("Two sold same-category items in single sale", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		categories, err := queries.GetCategories(db)
		require.NoError(t, err)

		totals := createTotalMap(categories)

		seller := setup.Seller()
		item1 := setup.Item(seller.UserId, aux.WithItemCategory(categories[0].CategoryId), aux.WithDummyData(1), aux.WithHidden(false))
		item2 := setup.Item(seller.UserId, aux.WithItemCategory(categories[0].CategoryId), aux.WithDummyData(2), aux.WithHidden(false))
		totals[item1.CategoryId] += item1.PriceInCents
		totals[item2.CategoryId] += item2.PriceInCents

		cashier := setup.Cashier()
		setup.Sale(cashier.UserId, []models.Id{item1.ItemId, item2.ItemId})

		categorySaleTotals, err := queries.GetSalesOverview(db)
		t.Log(categorySaleTotals)
		require.NoError(t, err)
		require.Equal(t, len(categories), len(categorySaleTotals))

		for categoryIndex, category := range categories {
			require.Equal(t, category.CategoryId, categorySaleTotals[categoryIndex].CategoryId)
			require.Equal(t, category.Name, categorySaleTotals[categoryIndex].CategoryName)
			require.Equal(t, totals[category.CategoryId], categorySaleTotals[categoryIndex].TotalInCents)
		}
	})

	t.Run("Two sold different-category items in single sale", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		categories, err := queries.GetCategories(db)
		require.NoError(t, err)

		totals := createTotalMap(categories)

		seller := setup.Seller()
		item1 := setup.Item(seller.UserId, aux.WithItemCategory(categories[0].CategoryId), aux.WithDummyData(1), aux.WithHidden(false))
		item2 := setup.Item(seller.UserId, aux.WithItemCategory(categories[0].CategoryId), aux.WithDummyData(2), aux.WithHidden(false))
		totals[item1.CategoryId] += item1.PriceInCents
		totals[item2.CategoryId] += item2.PriceInCents

		cashier := setup.Cashier()
		setup.Sale(cashier.UserId, []models.Id{item1.ItemId, item2.ItemId})

		categorySaleTotals, err := queries.GetSalesOverview(db)
		t.Log(categorySaleTotals)
		require.NoError(t, err)
		require.Equal(t, len(categories), len(categorySaleTotals))

		for categoryIndex, category := range categories {
			require.Equal(t, category.CategoryId, categorySaleTotals[categoryIndex].CategoryId)
			require.Equal(t, category.Name, categorySaleTotals[categoryIndex].CategoryName)
			require.Equal(t, totals[category.CategoryId], categorySaleTotals[categoryIndex].TotalInCents)
		}
	})

	t.Run("Two sold same-category items in separate sales", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		categories, err := queries.GetCategories(db)
		require.NoError(t, err)

		totals := createTotalMap(categories)

		seller := setup.Seller()
		item1 := setup.Item(seller.UserId, aux.WithItemCategory(categories[0].CategoryId), aux.WithDummyData(1), aux.WithHidden(false))
		item2 := setup.Item(seller.UserId, aux.WithItemCategory(categories[0].CategoryId), aux.WithDummyData(2), aux.WithHidden(false))
		totals[item1.CategoryId] += item1.PriceInCents
		totals[item2.CategoryId] += item2.PriceInCents

		cashier := setup.Cashier()
		setup.Sale(cashier.UserId, []models.Id{item1.ItemId})
		setup.Sale(cashier.UserId, []models.Id{item2.ItemId})

		categorySaleTotals, err := queries.GetSalesOverview(db)
		t.Log(categorySaleTotals)
		require.NoError(t, err)
		require.Equal(t, len(categories), len(categorySaleTotals))

		for categoryIndex, category := range categories {
			require.Equal(t, category.CategoryId, categorySaleTotals[categoryIndex].CategoryId)
			require.Equal(t, category.Name, categorySaleTotals[categoryIndex].CategoryName)
			require.Equal(t, totals[category.CategoryId], categorySaleTotals[categoryIndex].TotalInCents)
		}
	})

	t.Run("Two sold different-category items in separate sales", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		categories, err := queries.GetCategories(db)
		require.NoError(t, err)

		totals := createTotalMap(categories)

		seller := setup.Seller()
		item1 := setup.Item(seller.UserId, aux.WithItemCategory(categories[0].CategoryId), aux.WithDummyData(1), aux.WithHidden(false))
		item2 := setup.Item(seller.UserId, aux.WithItemCategory(categories[1].CategoryId), aux.WithDummyData(2), aux.WithHidden(false))
		totals[item1.CategoryId] += item1.PriceInCents
		totals[item2.CategoryId] += item2.PriceInCents

		cashier := setup.Cashier()
		setup.Sale(cashier.UserId, []models.Id{item1.ItemId})
		setup.Sale(cashier.UserId, []models.Id{item2.ItemId})

		categorySaleTotals, err := queries.GetSalesOverview(db)
		t.Log(categorySaleTotals)
		require.NoError(t, err)
		require.Equal(t, len(categories), len(categorySaleTotals))

		for categoryIndex, category := range categories {
			require.Equal(t, category.CategoryId, categorySaleTotals[categoryIndex].CategoryId)
			require.Equal(t, category.Name, categorySaleTotals[categoryIndex].CategoryName)
			require.Equal(t, totals[category.CategoryId], categorySaleTotals[categoryIndex].TotalInCents)
		}
	})
}
