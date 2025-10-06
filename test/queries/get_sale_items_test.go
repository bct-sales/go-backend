//go:build test

package queries

import (
	dberr "bctbackend/database/errors"
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
	itemIDs := []models.ID{
		setup.Item(seller.UserID, aux.WithDummyData(1), aux.WithHidden(false)).ItemID,
		setup.Item(seller.UserID, aux.WithDummyData(2), aux.WithHidden(false)).ItemID,
		setup.Item(seller.UserID, aux.WithDummyData(3), aux.WithHidden(false)).ItemID,
		setup.Item(seller.UserID, aux.WithDummyData(4), aux.WithHidden(false)).ItemID,
	}

	sale := setup.Sale(cashier.UserID, itemIDs)

	actualItems, err := queries.GetSaleItems(db, sale.SaleID)

	require.NoError(t, err)
	require.Len(t, actualItems, len(itemIDs))

	for index, actualItem := range actualItems {
		require.Equal(t, itemIDs[index], actualItem.ItemID)

		expectedItem, err := queries.GetItemWithID(db, itemIDs[index])

		require.NoError(t, err)
		require.Equal(t, expectedItem, actualItem)
	}
}

func TestGetSaleItemsOfNonexistentSale(t *testing.T) {
	setup, db := NewDatabaseFixture(WithDefaultCategories)
	defer setup.Close()

	saleID := models.ID(1)

	saleExists, err := queries.SaleWithIDExists(db, saleID)

	require.NoError(t, err)
	require.False(t, saleExists)

	_, err = queries.GetSaleItems(db, saleID)

	require.Error(t, err)
	requireDatabaseWrappedError(t, err, dberr.ErrNoSuchSale)
}
