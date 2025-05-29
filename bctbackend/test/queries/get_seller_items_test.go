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

func TestGetSellerItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("No items in database", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()

			actual, err := queries.GetSellerItems(db, seller.UserId, queries.AllItems)
			require.NoError(t, err)
			require.Empty(t, actual)
		})

		t.Run("Zero items associated with that seller", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			sellerWithoutItems := setup.Seller()
			sellerWithItems := setup.Seller()

			setup.Item(sellerWithItems.UserId, aux.WithDummyData(0), aux.WithHidden(false))
			setup.Item(sellerWithItems.UserId, aux.WithDummyData(1), aux.WithHidden(false))
			setup.Item(sellerWithItems.UserId, aux.WithDummyData(2), aux.WithHidden(false))
			setup.Item(sellerWithItems.UserId, aux.WithDummyData(3), aux.WithHidden(true))

			actual, err := queries.GetSellerItems(db, sellerWithoutItems.UserId, queries.AllItems)
			require.NoError(t, err)
			require.Empty(t, actual)
		})

		t.Run("Multiple items associated with seller, same timestamps", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()

			item1 := setup.Item(seller.UserId, aux.WithDummyData(0), aux.WithAddedAt(models.Timestamp(0)), aux.WithFrozen(false), aux.WithHidden(false))
			item2 := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithAddedAt(models.Timestamp(0)), aux.WithFrozen(false), aux.WithHidden(true))
			item3 := setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithAddedAt(models.Timestamp(0)), aux.WithFrozen(false), aux.WithHidden(true))
			item4 := setup.Item(seller.UserId, aux.WithDummyData(3), aux.WithAddedAt(models.Timestamp(0)), aux.WithFrozen(false), aux.WithHidden(false))

			expected := []*models.Item{item1, item2, item3, item4}
			actual, err := queries.GetSellerItems(db, seller.UserId, queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, expected, actual)
		})

		t.Run("Multiple items associated with seller, different timestamps", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()

			item1 := setup.Item(seller.UserId, aux.WithDummyData(0), aux.WithAddedAt(models.Timestamp(4)), aux.WithFrozen(false), aux.WithHidden(true))
			item2 := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithAddedAt(models.Timestamp(3)), aux.WithFrozen(false), aux.WithHidden(false))
			item3 := setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithAddedAt(models.Timestamp(2)), aux.WithFrozen(false), aux.WithHidden(false))
			item4 := setup.Item(seller.UserId, aux.WithDummyData(3), aux.WithAddedAt(models.Timestamp(1)), aux.WithFrozen(false), aux.WithHidden(false))

			expected := []*models.Item{item4, item3, item2, item1}
			actual, err := queries.GetSellerItems(db, seller.UserId, queries.AllItems)
			require.NoError(t, err)
			require.Equal(t, expected, actual)
		})

		t.Run("Only visible items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()

			items := setup.Items(seller.UserId, 20, aux.WithFrozen(false), aux.WithHidden(false), aux.WithAddedAt(models.Timestamp(0)))
			setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(true), aux.WithAddedAt(models.Timestamp(0)))

			actual, err := queries.GetSellerItems(db, seller.UserId, queries.OnlyVisibleItems)
			require.NoError(t, err)
			require.Equal(t, items, actual)
		})

		t.Run("Only hidden items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()

			setup.Items(seller.UserId, 20, aux.WithFrozen(false), aux.WithHidden(false), aux.WithAddedAt(models.Timestamp(0)))
			items := setup.Items(seller.UserId, 10, aux.WithFrozen(false), aux.WithHidden(true), aux.WithAddedAt(models.Timestamp(0)))

			actual, err := queries.GetSellerItems(db, seller.UserId, queries.OnlyHiddenItems)
			require.NoError(t, err)
			require.Equal(t, items, actual)
		})
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Unknown seller", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			unknownSellerId := models.Id(9999)
			setup.RequireNoSuchUsers(t, unknownSellerId)

			_, err := queries.GetSellerItems(db, unknownSellerId, queries.AllItems)
			require.ErrorIs(t, err, queries.ErrNoSuchUser)
		})

		t.Run("Wrong role: cashier", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier()

			_, err := queries.GetSellerItems(db, cashier.UserId, queries.AllItems)
			require.ErrorIs(t, err, queries.ErrInvalidRole)
		})

		t.Run("Wrong role: admin", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			admin := setup.Admin()

			_, err := queries.GetSellerItems(db, admin.UserId, queries.AllItems)
			require.ErrorIs(t, err, queries.ErrInvalidRole)
		})
	})
}
