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

func TestRemoveSale(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	sellerId := AddSellerToDatabase(db).UserId
	cashierId := AddCashierToDatabase(db).UserId
	sale1ItemIds := []models.Id{
		AddItemToDatabase(db, sellerId, WithDummyData(1)).ItemId,
		AddItemToDatabase(db, sellerId, WithDummyData(2)).ItemId,
	}
	sale2ItemIds := []models.Id{
		AddItemToDatabase(db, sellerId, WithDummyData(3)).ItemId,
		AddItemToDatabase(db, sellerId, WithDummyData(4)).ItemId,
	}

	sale1Id := AddSaleToDatabase(db, cashierId, sale1ItemIds)
	sale2Id := AddSaleToDatabase(db, cashierId, sale2ItemIds)

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
	db := OpenInitializedDatabase()
	defer db.Close()

	err := queries.RemoveSale(db, 0)

	var noSuchSaleError *queries.NoSuchSaleError
	require.ErrorAs(t, err, &noSuchSaleError)
}
