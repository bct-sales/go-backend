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

func TestAddSale(t *testing.T) {
	for _, itemIndices := range [][]int{{0}, {1}, {2}, {3}, {0, 1}, {1, 2, 3}, {0, 1, 2, 3}} {
		setup, db := NewDatabaseFixture()
		defer setup.Close()

		seller := setup.Seller()
		cashier := setup.Cashier()

		itemIds := []models.Id{
			setup.Item(seller.UserId, aux.WithDummyData(1)).ItemId,
			setup.Item(seller.UserId, aux.WithDummyData(2)).ItemId,
			setup.Item(seller.UserId, aux.WithDummyData(3)).ItemId,
			setup.Item(seller.UserId, aux.WithDummyData(4)).ItemId,
		}

		saleItemIds := make([]models.Id, len(itemIndices))
		for index, itemIndex := range itemIndices {
			saleItemIds[index] = itemIds[itemIndex]
		}

		timestamp := models.NewTimestamp(0)

		saleId, err := queries.AddSale(db, cashier.UserId, timestamp, saleItemIds)
		require.NoError(t, err)

		actualItems, err := queries.GetSaleItems(db, saleId)
		require.NoError(t, err)
		require.Len(t, actualItems, len(saleItemIds))

		for index, actualItem := range actualItems {
			require.Equal(t, saleItemIds[index], actualItem.ItemId)

			expectedItem, err := queries.GetItemWithId(db, saleItemIds[index])
			require.NoError(t, err)
			require.Equal(t, *expectedItem, actualItem)
		}
	}
}

func TestAddSaleWithoutItems(t *testing.T) {
	setup, db := NewDatabaseFixture()
	defer setup.Close()

	cashier := setup.Cashier()
	timestamp := models.NewTimestamp(0)

	_, err := queries.AddSale(db, cashier.UserId, timestamp, []models.Id{})
	require.Error(t, err)
}

func TestAddSaleWithSellerInsteadOfCashier(t *testing.T) {
	setup, db := NewDatabaseFixture()
	defer setup.Close()

	seller := setup.Seller()
	timestamp := models.NewTimestamp(0)
	itemId := setup.Item(seller.UserId, aux.WithDummyData(1)).ItemId

	_, err := queries.AddSale(db, seller.UserId, timestamp, []models.Id{itemId})
	require.Error(t, err)
}

func TestAddSaleWithSameItemTwice(t *testing.T) {
	setup, db := NewDatabaseFixture()
	defer setup.Close()

	seller := setup.Seller()
	timestamp := models.NewTimestamp(0)
	item := setup.Item(seller.UserId, aux.WithDummyData(1))

	_, err := queries.AddSale(db, seller.UserId, timestamp, []models.Id{item.ItemId, item.ItemId})
	var duplicateItemError *queries.DuplicateItemInSaleError
	require.ErrorAs(t, err, &duplicateItemError)
}
