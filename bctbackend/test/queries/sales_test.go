//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/test"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetSales(t *testing.T) {
	db := test.OpenInitializedDatabase()
	defer db.Close()

	sellerId := test.AddSellerToDatabase(db).UserId
	cashierId := test.AddCashierToDatabase(db).UserId

	itemIds := []models.Id{
		test.AddItemToDatabase(db, sellerId, 1).ItemId,
		test.AddItemToDatabase(db, sellerId, 2).ItemId,
		test.AddItemToDatabase(db, sellerId, 3).ItemId,
		test.AddItemToDatabase(db, sellerId, 4).ItemId,
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

func TestSaleExists(t *testing.T) {
	db := test.OpenInitializedDatabase()
	defer db.Close()

	sellerId := test.AddSellerToDatabase(db).UserId
	cashierId := test.AddCashierToDatabase(db).UserId
	itemId := test.AddItemToDatabase(db, sellerId, 1).ItemId

	saleId := test.AddSaleToDatabase(db, cashierId, []models.Id{itemId})
	saleExists, err := queries.SaleExists(db, saleId)
	require.NoError(t, err)
	require.True(t, saleExists)
}
