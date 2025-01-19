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

func TestSaleExists(t *testing.T) {
	db := setup.OpenInitializedDatabase()
	defer db.Close()

	sellerId := setup.AddSellerToDatabase(db).UserId
	cashierId := setup.AddCashierToDatabase(db).UserId
	itemId := setup.AddItemToDatabase(db, sellerId, setup.WithDummyData(1)).ItemId

	saleId := test.AddSaleToDatabase(db, cashierId, []models.Id{itemId})
	saleExists, err := queries.SaleExists(db, saleId)
	require.NoError(t, err)
	require.True(t, saleExists)
}
