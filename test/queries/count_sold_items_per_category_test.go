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

func TestCountSoldItemsPerCategory(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("No sold items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t)
			defer setup.Close()

			categoryID1 := models.ID(1)
			categoryID2 := models.ID(2)

			seller := setup.Seller(t)
			setup.Category(t, categoryID1, "foo")
			setup.Category(t, categoryID2, "bar")
			setup.Items(t, seller.UserID, 10, aux.WithItemCategory(categoryID1), aux.WithHidden(false))

			actualCounts, err := queries.CountSoldItemsPerCategory(t.Context(), db)
			require.NoError(t, err)
			require.Equal(t, len(actualCounts), 2)
			require.Equal(t, 0, actualCounts[categoryID1])
			require.Equal(t, 0, actualCounts[categoryID2])
		})

		t.Run("One sold item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t)
			defer setup.Close()

			categoryID1 := models.ID(1)
			categoryID2 := models.ID(2)

			seller := setup.Seller(t)
			cashier := setup.Cashier(t)
			setup.Category(t, categoryID1, "foo")
			setup.Category(t, categoryID2, "bar")
			setup.Items(t, seller.UserID, 10, aux.WithItemCategory(categoryID1), aux.WithHidden(false))
			soldItem := setup.Item(t, seller.UserID, aux.WithHidden(false), aux.WithItemCategory(categoryID1))
			setup.Sale(t, cashier.UserID, []models.ID{soldItem.ItemID})

			actualCounts, err := queries.CountSoldItemsPerCategory(t.Context(), db)
			require.NoError(t, err)
			require.Equal(t, len(actualCounts), 2)
			require.Equal(t, 1, actualCounts[categoryID1])
			require.Equal(t, 0, actualCounts[categoryID2])
		})

		t.Run("Two sold items in same category", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t)
			defer setup.Close()

			categoryID1 := models.ID(1)
			categoryID2 := models.ID(2)

			seller := setup.Seller(t)
			cashier := setup.Cashier(t)
			setup.Category(t, categoryID1, "foo")
			setup.Category(t, categoryID2, "bar")
			setup.Items(t, seller.UserID, 10, aux.WithItemCategory(categoryID1), aux.WithHidden(false))
			soldItem1 := setup.Item(t, seller.UserID, aux.WithHidden(false), aux.WithItemCategory(categoryID1))
			soldItem2 := setup.Item(t, seller.UserID, aux.WithHidden(false), aux.WithItemCategory(categoryID1))
			setup.Sale(t, cashier.UserID, []models.ID{soldItem1.ItemID, soldItem2.ItemID})

			actualCounts, err := queries.CountSoldItemsPerCategory(t.Context(), db)
			require.NoError(t, err)
			require.Equal(t, len(actualCounts), 2)
			require.Equal(t, 2, actualCounts[categoryID1])
			require.Equal(t, 0, actualCounts[categoryID2])
		})

		t.Run("Two sold items in different categories", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t)
			defer setup.Close()

			categoryID1 := models.ID(1)
			categoryID2 := models.ID(2)
			categoryID3 := models.ID(3)

			seller := setup.Seller(t)
			cashier := setup.Cashier(t)
			setup.Category(t, categoryID1, "foo")
			setup.Category(t, categoryID2, "bar")
			setup.Category(t, categoryID3, "baz")
			setup.Items(t, seller.UserID, 10, aux.WithItemCategory(categoryID1), aux.WithHidden(false))
			soldItem1 := setup.Item(t, seller.UserID, aux.WithHidden(false), aux.WithItemCategory(categoryID1))
			soldItem2 := setup.Item(t, seller.UserID, aux.WithHidden(false), aux.WithItemCategory(categoryID2))
			setup.Sale(t, cashier.UserID, []models.ID{soldItem1.ItemID, soldItem2.ItemID})

			actualCounts, err := queries.CountSoldItemsPerCategory(t.Context(), db)
			require.NoError(t, err)
			require.Equal(t, len(actualCounts), 3)
			require.Equal(t, 1, actualCounts[categoryID1])
			require.Equal(t, 1, actualCounts[categoryID2])
			require.Equal(t, 0, actualCounts[categoryID3])
		})
	})
}
