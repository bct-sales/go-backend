//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/test"
	"bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetSaleItems(t *testing.T) {
	db := setup.OpenInitializedDatabase()
	defer db.Close()

	sellerId := setup.AddSellerToDatabase(db).UserId
	cashierId := setup.AddCashierToDatabase(db).UserId
	itemIds := []models.Id{
		test.AddItemToDatabase(db, sellerId, test.WithDummyData(1)).ItemId,
		test.AddItemToDatabase(db, sellerId, test.WithDummyData(2)).ItemId,
		test.AddItemToDatabase(db, sellerId, test.WithDummyData(3)).ItemId,
		test.AddItemToDatabase(db, sellerId, test.WithDummyData(4)).ItemId,
	}

	saleId := test.AddSaleToDatabase(db, cashierId, itemIds)

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
	db := setup.OpenInitializedDatabase()
	defer db.Close()

	saleId := models.Id(1)

	saleExists, err := queries.SaleExists(db, saleId)

	require.NoError(t, err)
	require.False(t, saleExists)

	_, err = queries.GetSaleItems(db, saleId)

	require.Error(t, err)
	require.IsType(t, &queries.NoSuchSaleError{}, err)
}
