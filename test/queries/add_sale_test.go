//go:build test

package queries

import (
	"bctbackend/algorithms"
	dberr "bctbackend/database/errors"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"fmt"

	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddSale(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		for _, itemIndices := range [][]int{{0}, {1}, {2}, {3}, {0, 1}, {1, 2, 3}, {0, 1, 2, 3}, algorithms.Range(0, 10)} {
			testLabel := fmt.Sprintf("Item indices: %v", itemIndices)

			t.Run(testLabel, func(t *testing.T) {
				setup, db := NewDatabaseFixture(WithDefaultCategories)
				defer setup.Close()

				seller := setup.Seller()
				cashier := setup.Cashier()

				items := setup.Items(seller.UserID, 10, aux.WithHidden(false))
				itemIds := models.CollectItemIds(items)

				saleItemIds := make([]models.ID, len(itemIndices))
				for index, itemIndex := range itemIndices {
					saleItemIds[index] = itemIds[itemIndex]
				}

				timestamp := models.Timestamp(0)

				var saleId models.ID
				setup.WithTransaction(t, func(transaction *queries.TransactionalDatabaseQuerier) {
					var err error
					saleId, err = queries.AddSale(transaction, cashier.UserID, timestamp, saleItemIds)
					require.NoError(t, err)
				})

				actualItems, err := queries.GetSaleItems(db, saleId)
				require.NoError(t, err)
				require.Len(t, actualItems, len(saleItemIds))

				for index, actualItem := range actualItems {
					require.Equal(t, saleItemIds[index], actualItem.ItemID)

					expectedItem, err := queries.GetItemWithID(db, saleItemIds[index])
					require.NoError(t, err)
					require.Equal(t, expectedItem, actualItem)
				}
			})
		}
	})

	t.Run("Failure", func(t *testing.T) {
		t.Run("Sale with no items", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier()
			timestamp := models.Timestamp(0)

			setup.WithTransaction(t, func(transaction *queries.TransactionalDatabaseQuerier) {
				_, err := queries.AddSale(transaction, cashier.UserID, timestamp, []models.ID{})
				requireDatabaseWrappedError(t, err, dberr.ErrSaleMissingItems)
			})
		})

		t.Run("Sale with nonexistent items", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			cashier := setup.Cashier()
			timestamp := models.Timestamp(0)

			nonexistentItemId := models.ID(9999)
			setup.RequireNoSuchItems(t, nonexistentItemId)

			setup.WithTransaction(t, func(transaction *queries.TransactionalDatabaseQuerier) {
				_, err := queries.AddSale(transaction, cashier.UserID, timestamp, []models.ID{nonexistentItemId})
				requireDatabaseWrappedError(t, err, dberr.ErrNoSuchItem)
			})
		})

		t.Run("As seller", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			timestamp := models.Timestamp(0)
			itemId := setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithHidden(false)).ItemID

			setup.WithTransaction(t, func(transaction *queries.TransactionalDatabaseQuerier) {
				_, err := queries.AddSale(transaction, seller.UserID, timestamp, []models.ID{itemId})
				requireDatabaseWrappedError(t, err, dberr.ErrSaleRequiresCashier)
			})
		})

		t.Run("As admin", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			admin := setup.Admin()
			timestamp := models.Timestamp(0)
			itemId := setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithHidden(false)).ItemID

			setup.WithTransaction(t, func(transaction *queries.TransactionalDatabaseQuerier) {
				_, err := queries.AddSale(transaction, admin.UserID, timestamp, []models.ID{itemId})
				requireDatabaseWrappedError(t, err, dberr.ErrSaleRequiresCashier)
			})
		})

		t.Run("Duplicate item in sale", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			timestamp := models.Timestamp(0)
			item := setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithHidden(false))

			setup.WithTransaction(t, func(transaction *queries.TransactionalDatabaseQuerier) {
				_, err := queries.AddSale(transaction, cashier.UserID, timestamp, []models.ID{item.ItemID, item.ItemID})
				requireDatabaseWrappedError(t, err, dberr.ErrDuplicateItemInSale)
			})
		})

		t.Run("Hidden item in sale", func(t *testing.T) {
			setup, _ := NewDatabaseFixture(WithDefaultCategories)
			defer setup.Close()

			seller := setup.Seller()
			cashier := setup.Cashier()
			timestamp := models.Timestamp(0)
			item := setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithHidden(true))

			setup.WithTransaction(t, func(transaction *queries.TransactionalDatabaseQuerier) {
				_, err := queries.AddSale(transaction, cashier.UserID, timestamp, []models.ID{item.ItemID})
				requireDatabaseWrappedError(t, err, dberr.ErrItemHidden)
			})
		})
	})
}
