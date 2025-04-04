//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test"
	aux "bctbackend/test/helpers"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetSellerItems(t *testing.T) {
	t.Run("No items in database", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		seller := setup.Seller()

		items, err := queries.GetSellerItems(db, seller.UserId)
		require.NoError(t, err)
		require.Empty(t, items)
	})

	t.Run("Zero items associated with that seller", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		sellerWithoutItems := setup.Seller()
		sellerWithItems := setup.Seller()

		setup.Item(sellerWithItems.UserId, aux.WithDummyData(0))
		setup.Item(sellerWithItems.UserId, aux.WithDummyData(1))
		setup.Item(sellerWithItems.UserId, aux.WithDummyData(2))
		setup.Item(sellerWithItems.UserId, aux.WithDummyData(3))

		items, err := queries.GetSellerItems(db, sellerWithoutItems.UserId)
		require.NoError(t, err)
		require.Empty(t, items)
	})

	t.Run("Multiple items associated with seller, same timestamps", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		seller := setup.Seller()

		item1 := setup.Item(seller.UserId, aux.WithDummyData(0), aux.WithAddedAt(models.NewTimestamp(0)))
		item2 := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithAddedAt(models.NewTimestamp(0)))
		item3 := setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithAddedAt(models.NewTimestamp(0)))
		item4 := setup.Item(seller.UserId, aux.WithDummyData(3), aux.WithAddedAt(models.NewTimestamp(0)))

		items, err := queries.GetSellerItems(db, seller.UserId)
		require.NoError(t, err)
		require.Equal(t, []*models.Item{item1, item2, item3, item4}, items)
	})

	t.Run("Multiple items associated with seller, different timestamps", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		seller := setup.Seller()

		item1 := setup.Item(seller.UserId, aux.WithDummyData(0), aux.WithAddedAt(models.NewTimestamp(4)))
		item2 := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithAddedAt(models.NewTimestamp(3)))
		item3 := setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithAddedAt(models.NewTimestamp(2)))
		item4 := setup.Item(seller.UserId, aux.WithDummyData(3), aux.WithAddedAt(models.NewTimestamp(1)))

		items, err := queries.GetSellerItems(db, seller.UserId)
		require.NoError(t, err)
		require.Equal(t, []*models.Item{item4, item3, item2, item1}, items)
	})

	t.Run("Unknown seller", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		unknownSellerId := models.Id(9999)

		{
			userExists, err := queries.UserWithIdExists(db, unknownSellerId)
			require.NoError(t, err)
			require.False(t, userExists)
		}

		_, err := queries.GetSellerItems(db, unknownSellerId)
		var noSuchUserError *queries.NoSuchUserError
		require.ErrorAs(t, err, &noSuchUserError)
	})

	t.Run("Wrong role: cashier", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		cashier := setup.Cashier()

		_, err := queries.GetSellerItems(db, cashier.UserId)
		var invalidRoleError *queries.InvalidRoleError
		require.ErrorAs(t, err, &invalidRoleError)
	})

	t.Run("Wrong role: admin", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		admin := setup.Admin()

		_, err := queries.GetSellerItems(db, admin.UserId)
		var invalidRoleError *queries.InvalidRoleError
		require.ErrorAs(t, err, &invalidRoleError)
	})
}
