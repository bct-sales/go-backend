//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSellerItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("No items in database", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)

			actual, err := queries.GetSellerItems(t.Context(), db, seller.UserID, queries.AllItems)
			require.NoError(t, err)
			require.Empty(t, actual)
		})

		t.Run("Zero items associated with that seller", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			sellerWithoutItems := setup.Seller(t)
			sellerWithItems := setup.Seller(t)

			setup.Item(t, sellerWithItems.UserID, aux.WithDummyData(0), aux.WithHidden(false))
			setup.Item(t, sellerWithItems.UserID, aux.WithDummyData(1), aux.WithHidden(false))
			setup.Item(t, sellerWithItems.UserID, aux.WithDummyData(2), aux.WithHidden(false))
			setup.Item(t, sellerWithItems.UserID, aux.WithDummyData(3), aux.WithHidden(true))

			actual, err := queries.GetSellerItems(t.Context(), db, sellerWithoutItems.UserID, queries.AllItems)
			require.NoError(t, err)
			require.Empty(t, actual)
		})

		t.Run("Multiple items associated with seller, same timestamps", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)

			item1 := setup.Item(t, seller.UserID, aux.WithDummyData(0), aux.WithAddedAt(models.Timestamp(0)), aux.WithFrozen(false), aux.WithHidden(false))
			item2 := setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithAddedAt(models.Timestamp(0)), aux.WithFrozen(false), aux.WithHidden(true))
			item3 := setup.Item(t, seller.UserID, aux.WithDummyData(2), aux.WithAddedAt(models.Timestamp(0)), aux.WithFrozen(false), aux.WithHidden(true))
			item4 := setup.Item(t, seller.UserID, aux.WithDummyData(3), aux.WithAddedAt(models.Timestamp(0)), aux.WithFrozen(false), aux.WithHidden(false))

			expected := []*models.Item{item1, item2, item3, item4}
			actual, err := queries.GetSellerItems(t.Context(), db, seller.UserID, queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, expected, actual)
		})

		t.Run("Multiple items associated with seller, different timestamps", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)

			item1 := setup.Item(t, seller.UserID, aux.WithDummyData(0), aux.WithAddedAt(models.Timestamp(4)), aux.WithFrozen(false), aux.WithHidden(true))
			item2 := setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithAddedAt(models.Timestamp(3)), aux.WithFrozen(false), aux.WithHidden(false))
			item3 := setup.Item(t, seller.UserID, aux.WithDummyData(2), aux.WithAddedAt(models.Timestamp(2)), aux.WithFrozen(false), aux.WithHidden(false))
			item4 := setup.Item(t, seller.UserID, aux.WithDummyData(3), aux.WithAddedAt(models.Timestamp(1)), aux.WithFrozen(false), aux.WithHidden(false))

			expected := []*models.Item{item4, item3, item2, item1}
			actual, err := queries.GetSellerItems(t.Context(), db, seller.UserID, queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, expected, actual)
		})

		t.Run("Only visible items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)

			items := setup.Items(t, seller.UserID, 20, aux.WithFrozen(false), aux.WithHidden(false), aux.WithAddedAt(models.Timestamp(0)))
			setup.Items(t, seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(true), aux.WithAddedAt(models.Timestamp(0)))

			actual, err := queries.GetSellerItems(t.Context(), db, seller.UserID, queries.OnlyVisibleItems)
			require.NoError(t, err)
			require.Equal(t, items, actual)
		})

		t.Run("Only hidden items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller(t)

			setup.Items(t, seller.UserID, 20, aux.WithFrozen(false), aux.WithHidden(false), aux.WithAddedAt(models.Timestamp(0)))
			items := setup.Items(t, seller.UserID, 10, aux.WithFrozen(false), aux.WithHidden(true), aux.WithAddedAt(models.Timestamp(0)))

			actual, err := queries.GetSellerItems(t.Context(), db, seller.UserID, queries.OnlyHiddenItems)
			require.NoError(t, err)
			require.Equal(t, items, actual)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Unknown seller", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			unknownSellerID := setup.GenerateNonexistentUserID(t)

			_, err := queries.GetSellerItems(t.Context(), db, unknownSellerID, queries.AllItems)
			requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
		})

		t.Run("Wrong role: cashier", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier(t)

			_, err := queries.GetSellerItems(t.Context(), db, cashier.UserID, queries.AllItems)
			requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
		})

		t.Run("Wrong role: admin", func(t *testing.T) {
			setup, db := NewDatabaseFixture(t, WithDefaultCategories)
			defer setup.Close()

			admin := setup.Admin(t)

			_, err := queries.GetSellerItems(t.Context(), db, admin.UserID, queries.AllItems)
			requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
		})
	})
}
