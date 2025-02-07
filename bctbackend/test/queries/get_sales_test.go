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

func TestGetSales(t *testing.T) {
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

	saleIds := make([]models.Id, len(itemIds))
	for _, itemId := range itemIds {
		AddSaleToDatabase(db, cashierId, []models.Id{itemId})
	}

	actualSales := []*models.SaleSummary{}
	err := queries.GetSales(db, queries.CollectTo(&actualSales))
	require.NoError(t, err)
	require.Len(t, actualSales, len(saleIds))

	for _, actualSale := range actualSales {
		require.Equal(t, cashierId, actualSale.CashierId)

		saleItems, err := queries.GetSaleItems(db, actualSale.SaleId)
		require.NoError(t, err)
		require.Equal(t, 1, len(saleItems))
	}
}
