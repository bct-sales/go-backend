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

func TestGetSoldItems(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		t.Run("No items in existence", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			soldItems, err := queries.GetSoldItems(db)
			require.NoError(t, err)
			require.Empty(t, soldItems)
		})

		t.Run("Single unsold item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))

			soldItems, err := queries.GetSoldItems(db)
			require.NoError(t, err)
			require.Empty(t, soldItems)
		})

		t.Run("Single sold item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.Id{item.ItemId})

			soldItems, err := queries.GetSoldItems(db)
			require.NoError(t, err)
			require.Len(t, soldItems, 1)
			require.Equal(t, item, soldItems[0])
		})

		t.Run("Doubly sold item", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.Id{item.ItemId})
			setup.Sale(cashier.UserId, []models.Id{item.ItemId})

			soldItems, err := queries.GetSoldItems(db)
			require.NoError(t, err)
			require.Len(t, soldItems, 1)
			require.Equal(t, item, soldItems[0])
		})

		t.Run("Two sold items in single sale", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			item1 := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
			item2 := setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.Id{item1.ItemId, item2.ItemId})

			soldItems, err := queries.GetSoldItems(db)
			require.NoError(t, err)
			require.Len(t, soldItems, 2)
			require.Equal(t, item1, soldItems[0])
			require.Equal(t, item2, soldItems[1])
		})

		t.Run("Two sold items in separate sales", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			item1 := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))
			item2 := setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithHidden(false))
			setup.Sale(cashier.UserId, []models.Id{item1.ItemId}, aux.WithTransactionTime(models.Timestamp(100)))
			setup.Sale(cashier.UserId, []models.Id{item2.ItemId}, aux.WithTransactionTime(models.Timestamp(200)))

			soldItems, err := queries.GetSoldItems(db)
			require.NoError(t, err)
			require.Len(t, soldItems, 2)
			require.Equal(t, item1, soldItems[1])
			require.Equal(t, item2, soldItems[0])
		})
	})
}
