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
}
