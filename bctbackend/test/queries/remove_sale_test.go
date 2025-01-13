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

func TestRemoveSale(t *testing.T) {
	db := test.OpenInitializedDatabase()
	defer db.Close()

	sellerId := test.AddSellerToDatabase(db).UserId
	cashierId := test.AddCashierToDatabase(db).UserId
	sale1ItemIds := []models.Id{
		test.AddItemToDatabase(db, sellerId, 1).ItemId,
		test.AddItemToDatabase(db, sellerId, 2).ItemId,
	}
	sale2ItemIds := []models.Id{
		test.AddItemToDatabase(db, sellerId, 3).ItemId,
		test.AddItemToDatabase(db, sellerId, 4).ItemId,
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
	db := test.OpenInitializedDatabase()
	defer db.Close()

	err := queries.RemoveSale(db, 0)

	var noSuchSaleError *queries.NoSuchSaleError
	require.ErrorAs(t, err, &noSuchSaleError)
}
