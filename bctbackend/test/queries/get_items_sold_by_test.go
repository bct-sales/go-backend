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

func TestGetItemsSoldBy(t *testing.T) {
	t.Run("Zero items sold", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()
		setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
		setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithHidden(false))
		setup.Item(seller.UserId, aux.WithDummyData(3), aux.WithHidden(false))
		setup.Item(seller.UserId, aux.WithDummyData(4), aux.WithHidden(false))

		items, err := queries.GetItemsSoldBy(db, cashier.UserId)
		require.NoError(t, err)
		require.Len(t, items, 0)
	})

	t.Run("Items sold by different cashier", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		zeroSaleCashier := setup.Cashier()
		cashierWithSales := setup.Cashier()

		itemIds := []models.Id{
			setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false)).ItemID,
			setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithHidden(false)).ItemID,
			setup.Item(seller.UserId, aux.WithDummyData(3), aux.WithHidden(false)).ItemID,
			setup.Item(seller.UserId, aux.WithDummyData(4), aux.WithHidden(false)).ItemID,
		}

		setup.Sale(cashierWithSales.UserId, itemIds)

		items, err := queries.GetItemsSoldBy(db, zeroSaleCashier.UserId)
		require.NoError(t, err)
		require.Len(t, items, 0)
	})

	t.Run("Four items in single sale", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()

		expectedItems := []*models.Item{
			setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false)),
			setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithHidden(false)),
			setup.Item(seller.UserId, aux.WithDummyData(3), aux.WithHidden(false)),
			setup.Item(seller.UserId, aux.WithDummyData(4), aux.WithHidden(false)),
		}

		setup.Sale(cashier.UserId, []models.Id{expectedItems[0].ItemID, expectedItems[1].ItemID, expectedItems[2].ItemID, expectedItems[3].ItemID})

		actualItems, err := queries.GetItemsSoldBy(db, cashier.UserId)
		require.NoError(t, err)
		require.Len(t, actualItems, 4)
		require.Equal(t, expectedItems, actualItems)
	})

	t.Run("Four items in single sale, reordered", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()

		item1 := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
		item2 := setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithHidden(false))
		item3 := setup.Item(seller.UserId, aux.WithDummyData(3), aux.WithHidden(false))
		item4 := setup.Item(seller.UserId, aux.WithDummyData(4), aux.WithHidden(false))

		expectedItems := []*models.Item{item1, item2, item3, item4}
		setup.Sale(cashier.UserId, []models.Id{item4.ItemID, item3.ItemID, item2.ItemID, item1.ItemID})

		actualItems, err := queries.GetItemsSoldBy(db, cashier.UserId)
		require.NoError(t, err)
		require.Len(t, expectedItems, 4)
		require.Equal(t, expectedItems, actualItems)
	})

	t.Run("Four items in separate sales", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()

		item1 := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
		item2 := setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithHidden(false))
		item3 := setup.Item(seller.UserId, aux.WithDummyData(3), aux.WithHidden(false))
		item4 := setup.Item(seller.UserId, aux.WithDummyData(4), aux.WithHidden(false))

		setup.Sale(cashier.UserId, []models.Id{item1.ItemID, item2.ItemID})
		setup.Sale(cashier.UserId, []models.Id{item3.ItemID, item4.ItemID})

		items, err := queries.GetItemsSoldBy(db, cashier.UserId)
		require.NoError(t, err)
		require.Len(t, items, 4)
		require.Equal(t, []*models.Item{item1, item2, item3, item4}, items)
	})

	t.Run("Four items in separate sales, reordered", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()

		item1 := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
		item2 := setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithHidden(false))
		item3 := setup.Item(seller.UserId, aux.WithDummyData(3), aux.WithHidden(false))
		item4 := setup.Item(seller.UserId, aux.WithDummyData(4), aux.WithHidden(false))

		setup.Sale(cashier.UserId, []models.Id{item2.ItemID, item1.ItemID}, aux.WithTransactionTime(models.Timestamp(1)))
		setup.Sale(cashier.UserId, []models.Id{item4.ItemID, item3.ItemID}, aux.WithTransactionTime(models.Timestamp(0)))

		items, err := queries.GetItemsSoldBy(db, cashier.UserId)
		require.NoError(t, err)
		require.Len(t, items, 4)
		require.Equal(t, []*models.Item{item1, item2, item3, item4}, items)
	})

	t.Run("Four items in separate sales, reordered 2", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()

		item1 := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
		item2 := setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithHidden(false))
		item3 := setup.Item(seller.UserId, aux.WithDummyData(3), aux.WithHidden(false))
		item4 := setup.Item(seller.UserId, aux.WithDummyData(4), aux.WithHidden(false))

		setup.Sale(cashier.UserId, []models.Id{item2.ItemID, item1.ItemID}, aux.WithTransactionTime(models.Timestamp(0)))
		setup.Sale(cashier.UserId, []models.Id{item4.ItemID, item3.ItemID}, aux.WithTransactionTime(models.Timestamp(1)))

		items, err := queries.GetItemsSoldBy(db, cashier.UserId)
		require.NoError(t, err)
		require.Len(t, items, 4)
		require.Equal(t, []*models.Item{item3, item4, item1, item2}, items)
	})

	t.Run("Cashier does not exist", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		cashier := setup.Cashier()
		unknownCashierId := cashier.UserId + 1
		setup.RequireNoSuchUsers(t, unknownCashierId)

		_, err := queries.GetItemsSoldBy(db, unknownCashierId)
		require.ErrorIs(t, err, dberr.ErrNoSuchUser)
	})

	t.Run("User has wrong role", func(t *testing.T) {
		setup, db := NewDatabaseFixture(WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller()

		_, err := queries.GetItemsSoldBy(db, seller.UserId)
		require.ErrorIs(t, err, dberr.ErrWrongRole)
	})
}
