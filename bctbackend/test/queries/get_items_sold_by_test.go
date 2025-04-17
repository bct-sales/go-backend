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

func TestGetItemsSoldBy(t *testing.T) {
	t.Run("Zero items sold", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()
		setup.Item(seller.UserId, aux.WithDummyData(1))
		setup.Item(seller.UserId, aux.WithDummyData(2))
		setup.Item(seller.UserId, aux.WithDummyData(3))
		setup.Item(seller.UserId, aux.WithDummyData(4))

		items, err := queries.GetItemsSoldBy(db, cashier.UserId)
		require.NoError(t, err)
		require.Len(t, items, 0)
	})

	t.Run("Items sold by different cashier", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		seller := setup.Seller()
		zeroSaleCashier := setup.Cashier()
		cashierWithSales := setup.Cashier()

		itemIds := []models.Id{
			setup.Item(seller.UserId, aux.WithDummyData(1)).ItemId,
			setup.Item(seller.UserId, aux.WithDummyData(2)).ItemId,
			setup.Item(seller.UserId, aux.WithDummyData(3)).ItemId,
			setup.Item(seller.UserId, aux.WithDummyData(4)).ItemId,
		}

		setup.Sale(cashierWithSales.UserId, itemIds)

		items, err := queries.GetItemsSoldBy(db, zeroSaleCashier.UserId)
		require.NoError(t, err)
		require.Len(t, items, 0)
	})

	t.Run("Four items in single sale", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()

		expectedItems := []*models.Item{
			setup.Item(seller.UserId, aux.WithDummyData(1)),
			setup.Item(seller.UserId, aux.WithDummyData(2)),
			setup.Item(seller.UserId, aux.WithDummyData(3)),
			setup.Item(seller.UserId, aux.WithDummyData(4)),
		}

		setup.Sale(cashier.UserId, []models.Id{expectedItems[0].ItemId, expectedItems[1].ItemId, expectedItems[2].ItemId, expectedItems[3].ItemId})

		actualItems, err := queries.GetItemsSoldBy(db, cashier.UserId)
		require.NoError(t, err)
		require.Len(t, actualItems, 4)
		require.Equal(t, expectedItems, actualItems)
	})

	t.Run("Four items in single sale, reordered", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()

		item1 := setup.Item(seller.UserId, aux.WithDummyData(1))
		item2 := setup.Item(seller.UserId, aux.WithDummyData(2))
		item3 := setup.Item(seller.UserId, aux.WithDummyData(3))
		item4 := setup.Item(seller.UserId, aux.WithDummyData(4))

		expectedItems := []*models.Item{item1, item2, item3, item4}
		setup.Sale(cashier.UserId, []models.Id{item4.ItemId, item3.ItemId, item2.ItemId, item1.ItemId})

		actualItems, err := queries.GetItemsSoldBy(db, cashier.UserId)
		require.NoError(t, err)
		require.Len(t, expectedItems, 4)
		require.Equal(t, expectedItems, actualItems)
	})

	t.Run("Four items in separate sales", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()

		item1 := setup.Item(seller.UserId, aux.WithDummyData(1))
		item2 := setup.Item(seller.UserId, aux.WithDummyData(2))
		item3 := setup.Item(seller.UserId, aux.WithDummyData(3))
		item4 := setup.Item(seller.UserId, aux.WithDummyData(4))

		setup.Sale(cashier.UserId, []models.Id{item1.ItemId, item2.ItemId})
		setup.Sale(cashier.UserId, []models.Id{item3.ItemId, item4.ItemId})

		items, err := queries.GetItemsSoldBy(db, cashier.UserId)
		require.NoError(t, err)
		require.Len(t, items, 4)
		require.Equal(t, []*models.Item{item1, item2, item3, item4}, items)
	})

	t.Run("Four items in separate sales, reordered", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()

		item1 := setup.Item(seller.UserId, aux.WithDummyData(1))
		item2 := setup.Item(seller.UserId, aux.WithDummyData(2))
		item3 := setup.Item(seller.UserId, aux.WithDummyData(3))
		item4 := setup.Item(seller.UserId, aux.WithDummyData(4))

		setup.Sale(cashier.UserId, []models.Id{item2.ItemId, item1.ItemId}, aux.WithTransactionTime(models.NewTimestamp(1)))
		setup.Sale(cashier.UserId, []models.Id{item4.ItemId, item3.ItemId}, aux.WithTransactionTime(models.NewTimestamp(0)))

		items, err := queries.GetItemsSoldBy(db, cashier.UserId)
		require.NoError(t, err)
		require.Len(t, items, 4)
		require.Equal(t, []*models.Item{item1, item2, item3, item4}, items)
	})

	t.Run("Four items in separate sales, reordered 2", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()

		item1 := setup.Item(seller.UserId, aux.WithDummyData(1))
		item2 := setup.Item(seller.UserId, aux.WithDummyData(2))
		item3 := setup.Item(seller.UserId, aux.WithDummyData(3))
		item4 := setup.Item(seller.UserId, aux.WithDummyData(4))

		setup.Sale(cashier.UserId, []models.Id{item2.ItemId, item1.ItemId}, aux.WithTransactionTime(models.NewTimestamp(0)))
		setup.Sale(cashier.UserId, []models.Id{item4.ItemId, item3.ItemId}, aux.WithTransactionTime(models.NewTimestamp(1)))

		items, err := queries.GetItemsSoldBy(db, cashier.UserId)
		require.NoError(t, err)
		require.Len(t, items, 4)
		require.Equal(t, []*models.Item{item3, item4, item1, item2}, items)
	})

	t.Run("Cashier does not exist", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		cashier := setup.Cashier()
		unknownCashierId := cashier.UserId + 1

		{
			userExists, err := queries.UserWithIdExists(db, unknownCashierId)
			require.NoError(t, err)
			require.False(t, userExists)
		}

		_, err := queries.GetItemsSoldBy(db, unknownCashierId)
		var noSuchUserError *queries.NoSuchUserError
		require.ErrorAs(t, err, &noSuchUserError)
	})

	t.Run("User has wrong role", func(t *testing.T) {
		setup, db := Setup()
		defer setup.Close()

		seller := setup.Seller()

		_, err := queries.GetItemsSoldBy(db, seller.UserId)
		var invalidRoleError *queries.InvalidRoleError
		require.ErrorAs(t, err, &invalidRoleError)
	})
}
