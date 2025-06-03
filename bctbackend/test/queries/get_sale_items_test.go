//go:build test

package queries

import (
	"bctbackend/database"
	"bctbackend/database/models"
	"bctbackend/database/queries"
	aux "bctbackend/test/helpers"
	. "bctbackend/test/setup"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSaleItems(t *testing.T) {
	setup, db := NewDatabaseFixture(WithDefaultCategories)
	defer setup.Close()

	seller := setup.Seller()
	cashier := setup.Cashier()
	itemIds := []models.Id{
		setup.Item(seller.UserId, aux.WithDummyData(1), aux.WithHidden(false)).ItemID,
		setup.Item(seller.UserId, aux.WithDummyData(2), aux.WithHidden(false)).ItemID,
		setup.Item(seller.UserId, aux.WithDummyData(3), aux.WithHidden(false)).ItemID,
		setup.Item(seller.UserId, aux.WithDummyData(4), aux.WithHidden(false)).ItemID,
	}

	saleId := setup.Sale(cashier.UserId, itemIds)

	actualItems, err := queries.GetSaleItems(db, saleId)

	require.NoError(t, err)
	require.Len(t, actualItems, len(itemIds))

	for index, actualItem := range actualItems {
		require.Equal(t, itemIds[index], actualItem.ItemID)

		expectedItem, err := queries.GetItemWithId(db, itemIds[index])

		require.NoError(t, err)
		require.Equal(t, *expectedItem, actualItem)
	}
}

func TestGetSaleItemsOfNonexistentSale(t *testing.T) {
	setup, db := NewDatabaseFixture(WithDefaultCategories)
	defer setup.Close()

	saleId := models.Id(1)

	saleExists, err := queries.SaleExists(db, saleId)

	require.NoError(t, err)
	require.False(t, saleExists)

	_, err = queries.GetSaleItems(db, saleId)

	require.Error(t, err)
	require.ErrorIs(t, err, database.ErrNoSuchSale)
}
