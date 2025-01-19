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

func TestRemoveSale(t *testing.T) {
	db := setup.OpenInitializedDatabase()
	defer db.Close()

	sellerId := setup.AddSellerToDatabase(db).UserId
	cashierId := setup.AddCashierToDatabase(db).UserId
	sale1ItemIds := []models.Id{
		test.AddItemToDatabase(db, sellerId, test.WithDummyData(1)).ItemId,
		test.AddItemToDatabase(db, sellerId, test.WithDummyData(2)).ItemId,
	}
	sale2ItemIds := []models.Id{
		test.AddItemToDatabase(db, sellerId, test.WithDummyData(3)).ItemId,
		test.AddItemToDatabase(db, sellerId, test.WithDummyData(4)).ItemId,
	}

	sale1Id := test.AddSaleToDatabase(db, cashierId, sale1ItemIds)
	sale2Id := test.AddSaleToDatabase(db, cashierId, sale2ItemIds)

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
	db := setup.OpenInitializedDatabase()
	defer db.Close()

	err := queries.RemoveSale(db, 0)

	var noSuchSaleError *queries.NoSuchSaleError
	require.ErrorAs(t, err, &noSuchSaleError)
}
