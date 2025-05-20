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

func TestGetSales(t *testing.T) {
	setup, db := NewDatabaseFixture(WithDefaultCategories)
	defer setup.Close()

	seller := setup.Seller()
	cashier := setup.Cashier()

	itemIds := []models.Id{
		setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false)).ItemId,
		setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithHidden(false)).ItemId,
		setup.Item(seller.UserId, aux.WithDummyData(3), aux.WithHidden(false)).ItemId,
		setup.Item(seller.UserId, aux.WithDummyData(4), aux.WithHidden(false)).ItemId,
	}

	saleIds := make([]models.Id, len(itemIds))
	for _, itemId := range itemIds {
		setup.Sale(cashier.UserId, []models.Id{itemId})
	}

	actualSales := []*models.SaleSummary{}
	err := queries.GetSales(db, queries.CollectTo(&actualSales))
	require.NoError(t, err)
	require.Len(t, actualSales, len(saleIds))

	for _, actualSale := range actualSales {
		require.Equal(t, cashier.UserId, actualSale.CashierId)

		saleItems, err := queries.GetSaleItems(db, actualSale.SaleId)
		require.NoError(t, err)
		require.Equal(t, 1, len(saleItems))
	}
}
