//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetSellerItems(t *testing.T) {
	t.Run("No items in database", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId

		items, err := queries.GetSellerItems(db, sellerId)
		require.NoError(t, err)
		require.Empty(t, items)
	})

	t.Run("Zero items associated with that seller", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerWithoutItemsId := AddSellerToDatabase(db).UserId
		sellerWithItemsId := AddSellerToDatabase(db).UserId

		AddItemToDatabase(db, sellerWithItemsId, WithDummyData(0))
		AddItemToDatabase(db, sellerWithItemsId, WithDummyData(1))
		AddItemToDatabase(db, sellerWithItemsId, WithDummyData(2))
		AddItemToDatabase(db, sellerWithItemsId, WithDummyData(3))

		items, err := queries.GetSellerItems(db, sellerWithoutItemsId)
		require.NoError(t, err)
		require.Empty(t, items)
	})

	t.Run("Multiple items associated with seller, same timestamps", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId

		item1 := AddItemToDatabase(db, sellerId, WithDummyData(0), WithAddedAt(models.NewTimestamp(0)))
		item2 := AddItemToDatabase(db, sellerId, WithDummyData(1), WithAddedAt(models.NewTimestamp(0)))
		item3 := AddItemToDatabase(db, sellerId, WithDummyData(2), WithAddedAt(models.NewTimestamp(0)))
		item4 := AddItemToDatabase(db, sellerId, WithDummyData(3), WithAddedAt(models.NewTimestamp(0)))

		items, err := queries.GetSellerItems(db, sellerId)
		require.NoError(t, err)
		require.Equal(t, []*models.Item{item1, item2, item3, item4}, items)
	})

	t.Run("Multiple items associated with seller, different timestamps", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		sellerId := AddSellerToDatabase(db).UserId

		item1 := AddItemToDatabase(db, sellerId, WithDummyData(0), WithAddedAt(models.NewTimestamp(4)))
		item2 := AddItemToDatabase(db, sellerId, WithDummyData(1), WithAddedAt(models.NewTimestamp(3)))
		item3 := AddItemToDatabase(db, sellerId, WithDummyData(2), WithAddedAt(models.NewTimestamp(2)))
		item4 := AddItemToDatabase(db, sellerId, WithDummyData(3), WithAddedAt(models.NewTimestamp(1)))

		items, err := queries.GetSellerItems(db, sellerId)
		require.NoError(t, err)
		require.Equal(t, []*models.Item{item4, item3, item2, item1}, items)
	})

	t.Run("Unknown seller", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

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
		db := OpenInitializedDatabase()
		defer db.Close()

		cashierId := AddCashierToDatabase(db).UserId

		_, err := queries.GetSellerItems(db, cashierId)
		var invalidRoleError *queries.InvalidRoleError
		require.ErrorAs(t, err, &invalidRoleError)
	})

	t.Run("Wrong role: admin", func(t *testing.T) {
		db := OpenInitializedDatabase()
		defer db.Close()

		adminId := AddAdminToDatabase(db).UserId

		_, err := queries.GetSellerItems(db, adminId)
		var invalidRoleError *queries.InvalidRoleError
		require.ErrorAs(t, err, &invalidRoleError)
	})
}
