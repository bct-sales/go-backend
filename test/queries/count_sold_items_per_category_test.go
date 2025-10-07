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
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			categoryID1 := models.ID(1)
			categoryID2 := models.ID(2)

			seller := setup.Seller()
			setup.Category(categoryID1, "foo")
			setup.Category(categoryID2, "bar")
			setup.Items(seller.UserID, 10, aux.WithItemCategory(categoryID1), aux.WithHidden(false))

			actualCounts, err := queries.CountSoldItemsPerCategory(db)
			require.NoError(t, err)
			require.Equal(t, len(actualCounts), 2)
			require.Equal(t, 0, actualCounts[categoryID1])
			require.Equal(t, 0, actualCounts[categoryID2])
		})

		t.Run("One sold item", func(t *testing.T) {
			setup, db := NewDatabaseFixture()
			defer setup.Close()

			categoryID1 := models.ID(1)
			categoryID2 := models.ID(2)

			seller := setup.Seller()
			cashier := setup.Cashier()
			setup.Category(categoryID1, "foo")
			setup.Category(categoryID2, "bar")
			setup.Items(seller.UserID, 10, aux.WithItemCategory(categoryID1), aux.WithHidden(false))
			soldItem := setup.Item(seller.UserID, aux.WithHidden(false), aux.WithItemCategory(categoryID1))
			setup.Sale(cashier.UserID, []models.ID{soldItem.ItemID})

			actualCounts, err := queries.CountSoldItemsPerCategory(db)
			require.NoError(t, err)
			require.Equal(t, len(actualCounts), 2)
			require.Equal(t, 1, actualCounts[categoryID1])
			require.Equal(t, 0, actualCounts[categoryID2])
		})
	})
}
