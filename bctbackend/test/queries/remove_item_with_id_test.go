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

func TestRemoveExistingItem(t *testing.T) {
	db := setup.OpenInitializedDatabase()
	defer db.Close()

	sellerId := setup.AddSellerToDatabase(db).UserId
	itemId := setup.AddItemToDatabase(db, sellerId, setup.WithDummyData(1)).ItemId

	err := queries.RemoveItemWithId(db, itemId)

	require.NoError(t, err)

	itemExists, err := queries.ItemWithIdExists(db, itemId)
	require.NoError(t, err)
	require.False(t, itemExists)
}

func TestRemoveNonexistingItem(t *testing.T) {
	db := setup.OpenInitializedDatabase()
	defer db.Close()

	itemId := models.NewId(1)

	err := queries.RemoveItemWithId(db, itemId)

	var itemNotFoundError *queries.ItemNotFoundError
	require.ErrorAs(t, err, &itemNotFoundError)
}

func TestRemoveSoldItem(t *testing.T) {
	db := setup.OpenInitializedDatabase()
	defer db.Close()

	sellerId := setup.AddSellerToDatabase(db).UserId
	cashierId := setup.AddCashierToDatabase(db).UserId
	itemId := setup.AddItemToDatabase(db, sellerId, setup.WithDummyData(1)).ItemId

	test.AddSaleToDatabase(db, cashierId, []models.Id{itemId})

	err := queries.RemoveItemWithId(db, itemId)
	require.Error(t, err)

	itemExists, err := queries.ItemWithIdExists(db, itemId)
	require.NoError(t, err)
	require.True(t, itemExists)
}
