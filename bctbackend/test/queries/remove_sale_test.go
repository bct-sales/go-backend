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

func TestRemoveSale(t *testing.T) {
	setup, db := NewDatabaseFixture()
	defer setup.Close()

	seller := setup.Seller()
	cashier := setup.Cashier()
	sale1ItemIds := []models.Id{
		setup.Item(seller.UserId, aux.WithDummyData(1)).ItemId,
		setup.Item(seller.UserId, aux.WithDummyData(2)).ItemId,
	}
	sale2ItemIds := []models.Id{
		setup.Item(seller.UserId, aux.WithDummyData(3)).ItemId,
		setup.Item(seller.UserId, aux.WithDummyData(4)).ItemId,
	}

	sale1Id := setup.Sale(cashier.UserId, sale1ItemIds)
	sale2Id := setup.Sale(cashier.UserId, sale2ItemIds)

	err := queries.RemoveSale(db, sale1Id)
	require.NoError(t, err)

	sale1Exists, err := queries.SaleExists(db, sale1Id)
	require.NoError(t, err)
	require.False(t, sale1Exists)

	sale2Exists, err := queries.SaleExists(db, sale2Id)
	require.NoError(t, err)
	require.True(t, sale2Exists)
}

func TestRemoveNonexistentSale(t *testing.T) {
	setup, db := NewDatabaseFixture()
	defer setup.Close()

	err := queries.RemoveSale(db, 0)

	var noSuchSaleError *queries.NoSuchSaleError
	require.ErrorAs(t, err, &noSuchSaleError)
}
