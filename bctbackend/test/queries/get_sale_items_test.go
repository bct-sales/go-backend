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

func TestGetSaleItems(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	sellerId := AddSellerToDatabase(db).UserId
	cashierId := AddCashierToDatabase(db).UserId
	itemIds := []models.Id{
		AddItemToDatabase(db, sellerId, WithDummyData(1)).ItemId,
		AddItemToDatabase(db, sellerId, WithDummyData(2)).ItemId,
		AddItemToDatabase(db, sellerId, WithDummyData(3)).ItemId,
		AddItemToDatabase(db, sellerId, WithDummyData(4)).ItemId,
	}

	saleId := AddSaleToDatabase(db, cashierId, itemIds)

	actualItems, err := queries.GetSaleItems(db, saleId)

	require.NoError(t, err)
	require.Len(t, actualItems, len(itemIds))

	for index, actualItem := range actualItems {
		require.Equal(t, itemIds[index], actualItem.ItemId)

		expectedItem, err := queries.GetItemWithId(db, itemIds[index])

		require.NoError(t, err)
		require.Equal(t, *expectedItem, actualItem)
	}
}

func TestGetSaleItemsOfNonexistentSale(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	saleId := models.Id(1)

	saleExists, err := queries.SaleExists(db, saleId)

	require.NoError(t, err)
	require.False(t, saleExists)

	_, err = queries.GetSaleItems(db, saleId)

	require.Error(t, err)
	require.IsType(t, &queries.NoSuchSaleError{}, err)
}
