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

func TestGetSales(t *testing.T) {
	db := setup.OpenInitializedDatabase()
	defer db.Close()

	sellerId := test.AddSellerToDatabase(db).UserId
	cashierId := test.AddCashierToDatabase(db).UserId

	itemIds := []models.Id{
		test.AddItemToDatabase(db, sellerId, test.WithDummyData(1)).ItemId,
		test.AddItemToDatabase(db, sellerId, test.WithDummyData(2)).ItemId,
		test.AddItemToDatabase(db, sellerId, test.WithDummyData(3)).ItemId,
		test.AddItemToDatabase(db, sellerId, test.WithDummyData(4)).ItemId,
	}

	saleIds := make([]models.Id, len(itemIds))
	for _, itemId := range itemIds {
		test.AddSaleToDatabase(db, cashierId, []models.Id{itemId})
	}

	actualSales, err := queries.GetSales(db)

	require.NoError(t, err)
	require.Len(t, actualSales, len(saleIds))

	for _, actualSale := range actualSales {
		require.Equal(t, cashierId, actualSale.CashierId)

		saleItems, err := queries.GetSaleItems(db, actualSale.SaleId)
		require.NoError(t, err)
		require.Equal(t, 1, len(saleItems))
	}
}
