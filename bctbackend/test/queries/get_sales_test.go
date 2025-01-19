//go:build test

package queries

import (
	"bctbackend/database/models"
	"bctbackend/database/queries"
	"bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestGetSales(t *testing.T) {
	db := setup.OpenInitializedDatabase()
	defer db.Close()

	sellerId := setup.AddSellerToDatabase(db).UserId
	cashierId := setup.AddCashierToDatabase(db).UserId

	itemIds := []models.Id{
		setup.AddItemToDatabase(db, sellerId, setup.WithDummyData(1)).ItemId,
		setup.AddItemToDatabase(db, sellerId, setup.WithDummyData(2)).ItemId,
		setup.AddItemToDatabase(db, sellerId, setup.WithDummyData(3)).ItemId,
		setup.AddItemToDatabase(db, sellerId, setup.WithDummyData(4)).ItemId,
	}

	saleIds := make([]models.Id, len(itemIds))
	for _, itemId := range itemIds {
		setup.AddSaleToDatabase(db, cashierId, []models.Id{itemId})
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
