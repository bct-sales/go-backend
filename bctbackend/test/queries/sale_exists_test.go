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

func TestSaleExists(t *testing.T) {
	db := test.OpenInitializedDatabase()
	defer db.Close()

	sellerId := test.AddSellerToDatabase(db).UserId
	cashierId := test.AddCashierToDatabase(db).UserId
	itemId := test.AddItemToDatabase(db, sellerId, test.WithDummyData(1)).ItemId

	saleId := test.AddSaleToDatabase(db, cashierId, []models.Id{itemId})
	saleExists, err := queries.SaleExists(db, saleId)
	require.NoError(t, err)
	require.True(t, saleExists)
}
