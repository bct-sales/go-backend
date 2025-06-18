//go:build test

package queries

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"

	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddSale(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, itemIndices := range [][]int{{0}, {1}, {2}, {3}, {0, 1}, {1, 2, 3}, {0, 1, 2, 3}, algorithms.Range(0, 10)} {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()

			items := setup.Items(seller.UserId, 10, aux.WithHidden(false))
			itemIds := models.CollectItemIds(items)

			saleItemIds := make([]models.Id, len(itemIndices))
			for index, itemIndex := range itemIndices {
				saleItemIds[index] = itemIds[itemIndex]
			}

			timestamp := models.Timestamp(0)

			saleId, err := queries.AddSale(db, cashier.UserId, timestamp, saleItemIds)
			require.NoError(t, err)

			actualItems, err := queries.GetSaleItems(db, saleId)
			require.NoError(t, err)
			require.Len(t, actualItems, len(saleItemIds))

			for index, actualItem := range actualItems {
				require.Equal(t, saleItemIds[index], actualItem.ItemID)

				expectedItem, err := queries.GetItemWithId(db, saleItemIds[index])
				require.NoError(t, err)
				require.Equal(t, expectedItem, actualItem)
			}
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Sale with no items", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier()
			timestamp := models.Timestamp(0)

			_, err := queries.AddSale(db, cashier.UserId, timestamp, []models.Id{})
			require.Error(t, err)
		})

		t.Run("As seller", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			timestamp := models.Timestamp(0)
			itemId := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false)).ItemID

			_, err := queries.AddSale(db, seller.UserId, timestamp, []models.Id{itemId})
			require.Error(t, err)
		})

		t.Run("As admin", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			admin := setup.Admin()
			timestamp := models.Timestamp(0)
			itemId := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false)).ItemID

			_, err := queries.AddSale(db, admin.UserId, timestamp, []models.Id{itemId})
			require.Error(t, err)
		})

		t.Run("Duplicate item in sale", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			timestamp := models.Timestamp(0)
			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false))

			_, err := queries.AddSale(db, cashier.UserId, timestamp, []models.Id{item.ItemID, item.ItemID})
			require.ErrorIs(t, err, dberr.ErrDuplicateItemInSale)
		})

		t.Run("Hidden item in sale", func(t *testing.T) {
			setup, db := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			timestamp := models.Timestamp(0)
			item := setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(true))

			_, err := queries.AddSale(db, cashier.UserId, timestamp, []models.Id{item.ItemID})
			require.ErrorIs(t, err, dberr.ErrItemHidden)
		})
	})
}
