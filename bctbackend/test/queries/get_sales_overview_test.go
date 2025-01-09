//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/test"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func createTotalMap(categories []models.ItemCategory) map[models.Id]models.MoneyInCents {
	totalMap := make(map[models.Id]models.MoneyInCents)
	for _, category := range categories {
		totalMap[category.CategoryId] = models.MoneyInCents(0)
	}
	return totalMap
}

func TestGetSalesOverview(t *testing.T) {
	t.Run("Zero items", func(t *testing.T) {
		db := test.OpenInitializedDatabase()
		defer db.Close()

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
		db := test.OpenInitializedDatabase()
		defer db.Close()

		categories, err := queries.GetCategories(db)
		require.NoError(t, err)

		seller := test.AddSellerToDatabase(db)
		test.AddItemInCategoryToDatabase(db, seller.UserId, categories[0].CategoryId)

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
		db := test.OpenInitializedDatabase()
		defer db.Close()

		categories, err := queries.GetCategories(db)
		require.NoError(t, err)

		totals := createTotalMap(categories)

		seller := test.AddSellerToDatabase(db)
		item := test.AddItemInCategoryToDatabase(db, seller.UserId, categories[0].CategoryId)
		totals[item.CategoryId] += item.PriceInCents

		cashier := test.AddCashierToDatabase(db)
		test.AddSaleToDatabase(db, cashier.UserId, []models.Id{item.ItemId})

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
		db := test.OpenInitializedDatabase()
		defer db.Close()

		categories, err := queries.GetCategories(db)
		require.NoError(t, err)

		totals := createTotalMap(categories)

		seller := test.AddSellerToDatabase(db)
		item1 := test.AddItemInCategoryToDatabase(db, seller.UserId, categories[0].CategoryId)
		item2 := test.AddItemInCategoryToDatabase(db, seller.UserId, categories[0].CategoryId)
		totals[item1.CategoryId] += item1.PriceInCents
		totals[item2.CategoryId] += item1.PriceInCents

		cashier := test.AddCashierToDatabase(db)
		test.AddSaleToDatabase(db, cashier.UserId, []models.Id{item1.ItemId, item2.ItemId})

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
		db := test.OpenInitializedDatabase()
		defer db.Close()

		categories, err := queries.GetCategories(db)
		require.NoError(t, err)

		totals := createTotalMap(categories)

		seller := test.AddSellerToDatabase(db)
		item1 := test.AddItemInCategoryToDatabase(db, seller.UserId, categories[0].CategoryId)
		item2 := test.AddItemInCategoryToDatabase(db, seller.UserId, categories[1].CategoryId)
		totals[item1.CategoryId] += item1.PriceInCents
		totals[item2.CategoryId] += item1.PriceInCents

		cashier := test.AddCashierToDatabase(db)
		test.AddSaleToDatabase(db, cashier.UserId, []models.Id{item1.ItemId, item2.ItemId})

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
		db := test.OpenInitializedDatabase()
		defer db.Close()

		categories, err := queries.GetCategories(db)
		require.NoError(t, err)

		totals := createTotalMap(categories)

		seller := test.AddSellerToDatabase(db)
		item1 := test.AddItemInCategoryToDatabase(db, seller.UserId, categories[0].CategoryId)
		item2 := test.AddItemInCategoryToDatabase(db, seller.UserId, categories[0].CategoryId)
		totals[item1.CategoryId] += item1.PriceInCents
		totals[item2.CategoryId] += item1.PriceInCents

		cashier := test.AddCashierToDatabase(db)
		test.AddSaleToDatabase(db, cashier.UserId, []models.Id{item1.ItemId})
		test.AddSaleToDatabase(db, cashier.UserId, []models.Id{item2.ItemId})

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
		db := test.OpenInitializedDatabase()
		defer db.Close()

		categories, err := queries.GetCategories(db)
		require.NoError(t, err)

		totals := createTotalMap(categories)

		seller := test.AddSellerToDatabase(db)
		item1 := test.AddItemInCategoryToDatabase(db, seller.UserId, categories[0].CategoryId)
		item2 := test.AddItemInCategoryToDatabase(db, seller.UserId, categories[1].CategoryId)
		totals[item1.CategoryId] += item1.PriceInCents
		totals[item2.CategoryId] += item1.PriceInCents

		cashier := test.AddCashierToDatabase(db)
		test.AddSaleToDatabase(db, cashier.UserId, []models.Id{item1.ItemId})
		test.AddSaleToDatabase(db, cashier.UserId, []models.Id{item2.ItemId})

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
