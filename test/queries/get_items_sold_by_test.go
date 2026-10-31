//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetItemsSoldBy(t *testing.T) {
	t.Run("Zero items sold", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)
		cashier := setup.Cashier(t)
		setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))
		setup.Item(t, seller.UserID, aux.WithDummyData(2), aux.WithHidden(false))
		setup.Item(t, seller.UserID, aux.WithDummyData(3), aux.WithHidden(false))
		setup.Item(t, seller.UserID, aux.WithDummyData(4), aux.WithHidden(false))

		items, err := queries.GetItemsSoldBy(t.Context(), db, cashier.UserID)
		require.NoError(t, err)
		require.Len(t, items, 0)
	})

	t.Run("Single item", func(t *testing.T) {
		for i := range 20 {
			t.Run(fmt.Sprintf("i = %d", i), func(t *testing.T) {
				setup, db := NewDatabaseFixture(t, WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller(t)
				cashier := setup.Cashier(t)
				item := setup.Item(t, seller.UserID, aux.WithDummyData(i), aux.WithHidden(false))
				setup.Sale(t, cashier.UserID, []models.ID{item.ItemID})

				items, err := queries.GetItemsSoldBy(t.Context(), db, cashier.UserID)
				require.NoError(t, err)
				require.Len(t, items, 1)
				require.Equal(t, item, items[0])
			})
		}
	})

	t.Run("Items sold by different cashier", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)
		zeroSaleCashier := setup.Cashier(t)
		cashierWithSales := setup.Cashier(t)

		itemIDs := []models.ID{
			setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithHidden(false)).ItemID,
			setup.Item(t, seller.UserID, aux.WithDummyData(2), aux.WithHidden(false)).ItemID,
			setup.Item(t, seller.UserID, aux.WithDummyData(3), aux.WithHidden(false)).ItemID,
			setup.Item(t, seller.UserID, aux.WithDummyData(4), aux.WithHidden(false)).ItemID,
		}

		setup.Sale(t, cashierWithSales.UserID, itemIDs)

		items, err := queries.GetItemsSoldBy(t.Context(), db, zeroSaleCashier.UserID)
		require.NoError(t, err)
		require.Len(t, items, 0)
	})

	t.Run("Four items in single sale", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)
		cashier := setup.Cashier(t)

		expectedItems := []*models.Item{
			setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithHidden(false)),
			setup.Item(t, seller.UserID, aux.WithDummyData(2), aux.WithHidden(false)),
			setup.Item(t, seller.UserID, aux.WithDummyData(3), aux.WithHidden(false)),
			setup.Item(t, seller.UserID, aux.WithDummyData(4), aux.WithHidden(false)),
		}

		setup.Sale(t, cashier.UserID, []models.ID{expectedItems[0].ItemID, expectedItems[1].ItemID, expectedItems[2].ItemID, expectedItems[3].ItemID})

		actualItems, err := queries.GetItemsSoldBy(t.Context(), db, cashier.UserID)
		require.NoError(t, err)
		require.Len(t, actualItems, 4)
		require.Equal(t, expectedItems, actualItems)
	})

	t.Run("Four items in single sale, reordered", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)
		cashier := setup.Cashier(t)

		item1 := setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))
		item2 := setup.Item(t, seller.UserID, aux.WithDummyData(2), aux.WithHidden(false))
		item3 := setup.Item(t, seller.UserID, aux.WithDummyData(3), aux.WithHidden(false))
		item4 := setup.Item(t, seller.UserID, aux.WithDummyData(4), aux.WithHidden(false))

		expectedItems := []*models.Item{item1, item2, item3, item4}
		setup.Sale(t, cashier.UserID, []models.ID{item4.ItemID, item3.ItemID, item2.ItemID, item1.ItemID})

		actualItems, err := queries.GetItemsSoldBy(t.Context(), db, cashier.UserID)
		require.NoError(t, err)
		require.Len(t, expectedItems, 4)
		require.Equal(t, expectedItems, actualItems)
	})

	t.Run("Four items in separate sales", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)
		cashier := setup.Cashier(t)

		item1 := setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))
		item2 := setup.Item(t, seller.UserID, aux.WithDummyData(2), aux.WithHidden(false))
		item3 := setup.Item(t, seller.UserID, aux.WithDummyData(3), aux.WithHidden(false))
		item4 := setup.Item(t, seller.UserID, aux.WithDummyData(4), aux.WithHidden(false))

		setup.Sale(t, cashier.UserID, []models.ID{item1.ItemID, item2.ItemID})
		setup.Sale(t, cashier.UserID, []models.ID{item3.ItemID, item4.ItemID})

		items, err := queries.GetItemsSoldBy(t.Context(), db, cashier.UserID)
		require.NoError(t, err)
		require.Len(t, items, 4)
		require.Equal(t, []*models.Item{item1, item2, item3, item4}, items)
	})

	t.Run("Four items in separate sales, reordered", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)
		cashier := setup.Cashier(t)

		item1 := setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))
		item2 := setup.Item(t, seller.UserID, aux.WithDummyData(2), aux.WithHidden(false))
		item3 := setup.Item(t, seller.UserID, aux.WithDummyData(3), aux.WithHidden(false))
		item4 := setup.Item(t, seller.UserID, aux.WithDummyData(4), aux.WithHidden(false))

		setup.Sale(t, cashier.UserID, []models.ID{item2.ItemID, item1.ItemID}, aux.WithTransactionTime(models.Timestamp(1)))
		setup.Sale(t, cashier.UserID, []models.ID{item4.ItemID, item3.ItemID}, aux.WithTransactionTime(models.Timestamp(0)))

		items, err := queries.GetItemsSoldBy(t.Context(), db, cashier.UserID)
		require.NoError(t, err)
		require.Len(t, items, 4)
		require.Equal(t, []*models.Item{item1, item2, item3, item4}, items)
	})

	t.Run("Four items in separate sales, reordered 2", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)
		cashier := setup.Cashier(t)

		item1 := setup.Item(t, seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))
		item2 := setup.Item(t, seller.UserID, aux.WithDummyData(2), aux.WithHidden(false))
		item3 := setup.Item(t, seller.UserID, aux.WithDummyData(3), aux.WithHidden(false))
		item4 := setup.Item(t, seller.UserID, aux.WithDummyData(4), aux.WithHidden(false))

		setup.Sale(t, cashier.UserID, []models.ID{item2.ItemID, item1.ItemID}, aux.WithTransactionTime(models.Timestamp(0)))
		setup.Sale(t, cashier.UserID, []models.ID{item4.ItemID, item3.ItemID}, aux.WithTransactionTime(models.Timestamp(1)))

		items, err := queries.GetItemsSoldBy(t.Context(), db, cashier.UserID)
		require.NoError(t, err)
		require.Len(t, items, 4)
		require.Equal(t, []*models.Item{item3, item4, item1, item2}, items)
	})

	t.Run("Cashier does not exist", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		unknownCashierID := setup.GenerateNonexistentUserID(t)

		_, err := queries.GetItemsSoldBy(t.Context(), db, unknownCashierID)
		requireDatabaseWrappedError(t, err, dberr.ErrNoSuchUser)
	})

	t.Run("User has wrong role", func(t *testing.T) {
		setup, db := NewDatabaseFixture(t, WithDefaultCategories)
		defer setup.Close()

		seller := setup.Seller(t)

		_, err := queries.GetItemsSoldBy(t.Context(), db, seller.UserID)
		requireDatabaseWrappedError(t, err, dberr.ErrWrongRole)
	})
}
