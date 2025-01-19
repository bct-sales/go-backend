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

func TestSaleExists(t *testing.T) {
	db := OpenInitializedDatabase()
	defer db.Close()

	sellerId := AddSellerToDatabase(db).UserId
	cashierId := AddCashierToDatabase(db).UserId
	itemId := AddItemToDatabase(db, sellerId, WithDummyData(1)).ItemId

	saleId := AddSaleToDatabase(db, cashierId, []models.Id{itemId})
	saleExists, err := queries.SaleExists(db, saleId)
	require.NoError(t, err)
	require.True(t, saleExists)
}
